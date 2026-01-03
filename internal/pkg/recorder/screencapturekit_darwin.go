//go:build darwin
// +build darwin

package recorder

/*
#cgo CFLAGS: -x objective-c -mmacosx-version-min=13.0 -Wno-unguarded-availability-new
#cgo LDFLAGS: -framework ScreenCaptureKit -framework CoreMedia -framework CoreAudio -framework AudioToolbox -framework Foundation -framework AVFoundation -mmacosx-version-min=13.0

#import <ScreenCaptureKit/ScreenCaptureKit.h>
#import <CoreMedia/CoreMedia.h>
#import <AudioToolbox/AudioToolbox.h>
#import <AVFoundation/AVFoundation.h>
#include <stdlib.h>
#include <string.h>

static void* audioBufferData = NULL;
static size_t audioBufferSize = 0;
static size_t audioBufferCapacity = 0;
static bool isRecording = false;
static dispatch_queue_t audioQueue = NULL;
static NSLock* bufferLock = nil;

static const int kSampleRate = 48000;
static const int kChannels = 2;
static const int kBitsPerSample = 16;

API_AVAILABLE(macos(13.0))
@interface SCKAudioDelegate : NSObject <SCStreamOutput, SCStreamDelegate>
@property (nonatomic, strong) SCStream* stream;
@property (nonatomic, strong) SCContentFilter* filter;
@property (nonatomic, strong) SCStreamConfiguration* config;
@end

API_AVAILABLE(macos(13.0))
@implementation SCKAudioDelegate

- (void)stream:(SCStream *)stream didOutputSampleBuffer:(CMSampleBufferRef)sampleBuffer ofType:(SCStreamOutputType)type {
    if (@available(macOS 13.0, *)) {
        if (type != SCStreamOutputTypeAudio || !isRecording) {
            return;
        }
    }

    CMFormatDescriptionRef formatDesc = CMSampleBufferGetFormatDescription(sampleBuffer);
    const AudioStreamBasicDescription* asbd = CMAudioFormatDescriptionGetStreamBasicDescription(formatDesc);
    if (!asbd) {
        return;
    }

    CMBlockBufferRef blockBuffer = CMSampleBufferGetDataBuffer(sampleBuffer);
    if (!blockBuffer) {
        return;
    }

    size_t lengthAtOffset = 0;
    size_t totalLength = 0;
    char* dataPointer = NULL;

    OSStatus status = CMBlockBufferGetDataPointer(blockBuffer, 0, &lengthAtOffset, &totalLength, &dataPointer);
    if (status != kCMBlockBufferNoErr || !dataPointer || totalLength == 0) {
        return;
    }

    bool isNonInterleaved = (asbd->mFormatFlags & kAudioFormatFlagIsNonInterleaved) != 0;
    bool isFloat = (asbd->mFormatFlags & kAudioFormatFlagIsFloat) != 0;
    int channels = asbd->mChannelsPerFrame;

    size_t numFrames;
    if (isNonInterleaved) {
        numFrames = totalLength / (sizeof(float) * channels);
    } else {
        numFrames = totalLength / (sizeof(float) * channels);
    }
    size_t int16DataSize = numFrames * channels * sizeof(int16_t);

    [bufferLock lock];
    if (isRecording) {
        if (audioBufferSize + int16DataSize > audioBufferCapacity) {
            size_t newCapacity = (audioBufferCapacity == 0) ? 1024 * 1024 : audioBufferCapacity * 2;
            while (newCapacity < audioBufferSize + int16DataSize) {
                newCapacity *= 2;
            }
            void* newBuffer = realloc(audioBufferData, newCapacity);
            if (newBuffer) {
                audioBufferData = newBuffer;
                audioBufferCapacity = newCapacity;
            }
        }

        if (audioBufferSize + int16DataSize <= audioBufferCapacity) {
            int16_t* destBuffer = (int16_t*)((char*)audioBufferData + audioBufferSize);
            float* floatData = (float*)dataPointer;

            if (isNonInterleaved && channels == 2) {
                float* leftChannel = floatData;
                float* rightChannel = floatData + numFrames;
                for (size_t i = 0; i < numFrames; i++) {
                    float leftSample = leftChannel[i];
                    float rightSample = rightChannel[i];
                    if (leftSample > 1.0f) leftSample = 1.0f;
                    if (leftSample < -1.0f) leftSample = -1.0f;
                    if (rightSample > 1.0f) rightSample = 1.0f;
                    if (rightSample < -1.0f) rightSample = -1.0f;
                    destBuffer[i * 2] = (int16_t)(leftSample * 32767.0f);
                    destBuffer[i * 2 + 1] = (int16_t)(rightSample * 32767.0f);
                }
            } else {
                size_t numSamples = totalLength / sizeof(float);
                for (size_t i = 0; i < numSamples; i++) {
                    float sample = floatData[i];
                    if (sample > 1.0f) sample = 1.0f;
                    if (sample < -1.0f) sample = -1.0f;
                    destBuffer[i] = (int16_t)(sample * 32767.0f);
                }
            }
            audioBufferSize += int16DataSize;
        }
    }
    [bufferLock unlock];
}

- (void)stream:(SCStream *)stream didStopWithError:(NSError *)error {
    if (error) {
        NSLog(@"SCStream stopped with error: %@", error.localizedDescription);
    }
}

@end

static SCKAudioDelegate* audioDelegate API_AVAILABLE(macos(13.0)) = nil;

int sckIsSystemAudioSupported() {
    if (@available(macOS 13.0, *)) {
        return 1;
    }
    return 0;
}

int sckInitAudioCapture() {
    if (@available(macOS 13.0, *)) {
        if (bufferLock == nil) {
            bufferLock = [[NSLock alloc] init];
        }
        if (audioQueue == nil) {
            audioQueue = dispatch_queue_create("com.onx.screencapturekit.audio", DISPATCH_QUEUE_SERIAL);
        }
        return 1;
    }
    return 0;
}

int sckStartAudioCapture() {
    if (@available(macOS 13.0, *)) {
        __block int result = 0;
        dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);

        [SCShareableContent getShareableContentWithCompletionHandler:^(SCShareableContent* content, NSError* error) {
            if (error || !content) {
                NSLog(@"Failed to get shareable content: %@", error.localizedDescription);
                dispatch_semaphore_signal(semaphore);
                return;
            }

            SCDisplay* mainDisplay = content.displays.firstObject;
            if (!mainDisplay) {
                NSLog(@"No display found");
                dispatch_semaphore_signal(semaphore);
                return;
            }

            SCContentFilter* filter = [[SCContentFilter alloc] initWithDisplay:mainDisplay excludingWindows:@[]];

            SCStreamConfiguration* config = [[SCStreamConfiguration alloc] init];
            config.capturesAudio = YES;
            config.excludesCurrentProcessAudio = YES;
            config.sampleRate = kSampleRate;
            config.channelCount = kChannels;

            config.width = 2;
            config.height = 2;
            config.minimumFrameInterval = CMTimeMake(1, 1);
            config.showsCursor = NO;

            audioDelegate = [[SCKAudioDelegate alloc] init];
            audioDelegate.filter = filter;
            audioDelegate.config = config;

            SCStream* stream = [[SCStream alloc] initWithFilter:filter configuration:config delegate:audioDelegate];

            NSError* outputError = nil;
            [stream addStreamOutput:audioDelegate type:SCStreamOutputTypeAudio sampleHandlerQueue:audioQueue error:&outputError];
            if (outputError) {
                NSLog(@"Failed to add stream output: %@", outputError.localizedDescription);
                dispatch_semaphore_signal(semaphore);
                return;
            }

            audioDelegate.stream = stream;

            [bufferLock lock];
            audioBufferSize = 0;
            isRecording = true;
            [bufferLock unlock];

            [stream startCaptureWithCompletionHandler:^(NSError* startError) {
                if (startError) {
                    NSLog(@"Failed to start capture: %@", startError.localizedDescription);
                    [bufferLock lock];
                    isRecording = false;
                    [bufferLock unlock];
                } else {
                    result = 1;
                    NSLog(@"ScreenCaptureKit audio capture started successfully");
                }
                dispatch_semaphore_signal(semaphore);
            }];
        }];

        dispatch_semaphore_wait(semaphore, dispatch_time(DISPATCH_TIME_NOW, 5 * NSEC_PER_SEC));
        return result;
    }
    return 0;
}

int sckStopAudioCapture() {
    if (@available(macOS 13.0, *)) {
        [bufferLock lock];
        isRecording = false;
        [bufferLock unlock];

        if (audioDelegate && audioDelegate.stream) {
            dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);

            [audioDelegate.stream stopCaptureWithCompletionHandler:^(NSError* error) {
                if (error) {
                    NSLog(@"Error stopping capture: %@", error.localizedDescription);
                }
                dispatch_semaphore_signal(semaphore);
            }];

            dispatch_semaphore_wait(semaphore, dispatch_time(DISPATCH_TIME_NOW, 5 * NSEC_PER_SEC));
            audioDelegate.stream = nil;
        }

        NSLog(@"ScreenCaptureKit audio capture stopped, buffer size: %zu", audioBufferSize);
        return 1;
    }
    return 0;
}

size_t sckGetAudioBufferSize() {
    [bufferLock lock];
    size_t size = audioBufferSize;
    [bufferLock unlock];
    return size;
}

size_t sckCopyAudioBuffer(void* destBuffer, size_t destSize) {
    [bufferLock lock];
    size_t copySize = (destSize < audioBufferSize) ? destSize : audioBufferSize;
    if (copySize > 0 && destBuffer && audioBufferData) {
        memcpy(destBuffer, audioBufferData, copySize);
    }
    [bufferLock unlock];
    return copySize;
}

int sckGetSampleRate() {
    return kSampleRate;
}

int sckGetChannels() {
    return kChannels;
}

int sckGetBitsPerSample() {
    return kBitsPerSample;
}

void sckCleanup() {
    [bufferLock lock];
    if (audioBufferData) {
        free(audioBufferData);
        audioBufferData = NULL;
    }
    audioBufferSize = 0;
    audioBufferCapacity = 0;
    isRecording = false;
    [bufferLock unlock];

    if (@available(macOS 13.0, *)) {
        audioDelegate = nil;
    }
}

*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"unsafe"
)

// SCKAudioRecorder captures system audio using ScreenCaptureKit on macOS 13.0+
type SCKAudioRecorder struct {
	mu          sync.Mutex
	outputPath  string
	isRecording bool
	sampleRate  uint32
	channels    uint32
}

// IsSystemAudioSupported returns true if the system supports ScreenCaptureKit audio capture (macOS 13.0+)
func IsSystemAudioSupported() bool {
	return C.sckIsSystemAudioSupported() == 1
}

// NewSCKAudioRecorder creates a new ScreenCaptureKit-based audio recorder
func NewSCKAudioRecorder(outputPath string) (*SCKAudioRecorder, error) {
	if !IsSystemAudioSupported() {
		return nil, fmt.Errorf("system audio capture requires macOS 13.0 or later")
	}

	if C.sckInitAudioCapture() != 1 {
		return nil, fmt.Errorf("failed to initialize ScreenCaptureKit audio capture")
	}

	return &SCKAudioRecorder{
		outputPath: outputPath,
		sampleRate: uint32(C.sckGetSampleRate()),
		channels:   uint32(C.sckGetChannels()),
	}, nil
}

// Start begins capturing system audio
func (r *SCKAudioRecorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isRecording {
		return fmt.Errorf("already recording")
	}

	if C.sckStartAudioCapture() != 1 {
		return fmt.Errorf("failed to start ScreenCaptureKit audio capture")
	}

	r.isRecording = true
	return nil
}

// Stop stops capturing and writes the audio to a WAV file
func (r *SCKAudioRecorder) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isRecording {
		return fmt.Errorf("not recording")
	}

	r.isRecording = false

	// Stop the capture
	C.sckStopAudioCapture()

	// Get buffer size
	bufferSize := C.sckGetAudioBufferSize()
	if bufferSize == 0 {
		// No audio captured, write empty WAV
		return r.writeWAV(nil)
	}

	// Allocate Go buffer and copy data
	audioData := make([]byte, bufferSize)
	actualSize := C.sckCopyAudioBuffer(unsafe.Pointer(&audioData[0]), C.size_t(len(audioData)))
	audioData = audioData[:actualSize]

	// Clean up C resources
	C.sckCleanup()

	// Write WAV file
	return r.writeWAV(audioData)
}

// writeWAV writes the recorded samples to a WAV file
func (r *SCKAudioRecorder) writeWAV(samples []byte) error {
	file, err := os.Create(r.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %w", err)
	}
	defer file.Close()

	numSamples := uint32(len(samples))
	bitsPerSample := uint16(C.sckGetBitsPerSample())
	byteRate := r.sampleRate * uint32(r.channels) * uint32(bitsPerSample) / 8
	blockAlign := uint16(r.channels) * bitsPerSample / 8

	// Write RIFF header
	file.WriteString("RIFF")
	binary.Write(file, binary.LittleEndian, uint32(36+numSamples))
	file.WriteString("WAVE")

	// Write fmt subchunk
	file.WriteString("fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16))
	binary.Write(file, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(file, binary.LittleEndian, uint16(r.channels))
	binary.Write(file, binary.LittleEndian, r.sampleRate)
	binary.Write(file, binary.LittleEndian, byteRate)
	binary.Write(file, binary.LittleEndian, blockAlign)
	binary.Write(file, binary.LittleEndian, bitsPerSample)

	// Write data subchunk
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, numSamples)
	if len(samples) > 0 {
		file.Write(samples)
	}

	return nil
}
