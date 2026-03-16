export interface UploadHistoryItem {
    id: number;
    sessionId: string;
    filePath: string;
    filename: string;
    contentType: string;
    fileSize: number;
    status: string;
    attemptCount: number;
    lastError: string;
    lastHttpStatus: number;
    uploadId: string;
    createdAt: string;
    updatedAt: string;
    lastAttemptAt: string;
    nextAttemptAt: string;
    signedUrlExpiresAt: string;
    gcsUploadedAt: string;
    confirmedAt: string;
}

type UploadHistoryBridge = {
    GetUploadHistory: () => Promise<UploadHistoryItem[]>;
};

function getBridge(): UploadHistoryBridge {
    const bridge = (window as typeof window & {
        go?: {
            app?: {
                App?: UploadHistoryBridge;
            };
        };
    }).go?.app?.App;

    if (!bridge) {
        throw new Error('Wails bridge is not available');
    }

    return bridge;
}

export function GetUploadHistory(): Promise<UploadHistoryItem[]> {
    return getBridge().GetUploadHistory();
}
