//go:build darwin
// +build darwin

package activity

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework ApplicationServices

#import <Cocoa/Cocoa.h>
#import <ApplicationServices/ApplicationServices.h>

typedef struct {
    char* appName;
    char* bundleId;
    char* windowTitle;
    int32_t pid;
} ActiveWindowInfo;

ActiveWindowInfo getActiveWindow() {
    ActiveWindowInfo info = {NULL, NULL, NULL, 0};

    @autoreleasepool {
        NSRunningApplication* frontApp = [[NSWorkspace sharedWorkspace] frontmostApplication];
        if (frontApp) {
            info.pid = frontApp.processIdentifier;

            if (frontApp.localizedName) {
                info.appName = strdup([frontApp.localizedName UTF8String]);
            }

            if (frontApp.bundleIdentifier) {
                info.bundleId = strdup([frontApp.bundleIdentifier UTF8String]);
            }

            CFArrayRef windowList = CGWindowListCopyWindowInfo(
                kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements,
                kCGNullWindowID
            );

            if (windowList) {
                CFIndex count = CFArrayGetCount(windowList);
                for (CFIndex i = 0; i < count; i++) {
                    NSDictionary* windowInfo = (NSDictionary*)CFArrayGetValueAtIndex(windowList, i);
                    NSNumber* ownerPID = windowInfo[(NSString*)kCGWindowOwnerPID];

                    if ([ownerPID intValue] == info.pid) {
                        NSString* windowName = windowInfo[(NSString*)kCGWindowName];
                        NSNumber* windowLayer = windowInfo[(NSString*)kCGWindowLayer];

                        if (windowName && [windowName length] > 0 && [windowLayer intValue] == 0) {
                            info.windowTitle = strdup([windowName UTF8String]);
                            break;
                        }
                    }
                }
                CFRelease(windowList);
            }
        }
    }

    return info;
}

void freeWindowInfo(ActiveWindowInfo info) {
    if (info.appName) free(info.appName);
    if (info.bundleId) free(info.bundleId);
    if (info.windowTitle) free(info.windowTitle);
}

int64_t getIdleTime() {
    CFTimeInterval idleTime = CGEventSourceSecondsSinceLastEventType(
        kCGEventSourceStateCombinedSessionState,
        kCGAnyInputEventType
    );
    return (int64_t)(idleTime * 1000);
}
*/
import "C"

func GetActiveWindow() WindowInfo {
	cInfo := C.getActiveWindow()
	defer C.freeWindowInfo(cInfo)

	info := WindowInfo{
		ProcessID: int32(cInfo.pid),
	}

	if cInfo.appName != nil {
		info.AppName = C.GoString(cInfo.appName)
	}
	if cInfo.bundleId != nil {
		info.BundleID = C.GoString(cInfo.bundleId)
	}
	if cInfo.windowTitle != nil {
		info.WindowTitle = C.GoString(cInfo.windowTitle)
	}

	return info
}

func GetIdleTimeMs() int64 {
	return int64(C.getIdleTime())
}
