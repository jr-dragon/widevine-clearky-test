# Widevine DRM Demo

使用 W3C EME ClearKey 加密的 DASH 串流播放示範應用。

## 架構

```
原始 MP4 ──▶ Shaka Packager (CENC 加密) ──▶ DASH (MPD + segments)
                                                │
                Go Backend                      │
          ┌──────────────────────┐              │
          │ /api/encrypt  上傳加密 │              │
          │ /api/license  發放金鑰 │◀─────────────┘
          │ /api/videos   影片列表 │
          └──────────────────────┘
                    │
            Shaka Player (EME ClearKey)
```

- **加密方式**：CENC (AES-128-CTR)，與 Widevine 使用相同的加密格式
- **DRM Scheme**：W3C EME ClearKey（不需要第三方憑證）
- **打包格式**：MPEG-DASH
- **播放器**：[Shaka Player](https://github.com/shaka-project/shaka-player)
- **打包工具**：[Shaka Packager](https://github.com/shaka-project/shaka-packager)

## 快速開始

```bash
docker compose up --build
```

開啟瀏覽器前往 http://localhost:8080，透過網頁上傳 MP4 影片即可自動加密並播放。

## 本機開發

需要先安裝 [Shaka Packager](https://github.com/shaka-project/shaka-packager/releases)，並確保 `shaka-packager` 在 `$PATH` 中（或設定 `SHAKA_PACKAGER_BIN` 環境變數）。

```bash
go run .
```

## 環境變數

| 變數 | 預設值 | 說明 |
|------|--------|------|
| `ADDR` | `:8080` | HTTP 監聽地址 |
| `VIDEOS_DIR` | `./videos` | 影片儲存目錄 |
| `KEYS_PATH` | `./keys.json` | 金鑰儲存路徑 |
| `WEB_DIR` | `./web` | 前端靜態檔案目錄 |
| `SHAKA_PACKAGER_BIN` | `shaka-packager` | Shaka Packager 執行檔路徑 |

## API

### `POST /api/encrypt`

上傳 MP4 影片進行 CENC 加密與 DASH 打包。

- Content-Type: `multipart/form-data`
- 欄位: `video` (MP4 檔案)
- 回應: `{ "id": "...", "key_id": "...", "manifest": "/videos/.../manifest.mpd" }`

### `POST /api/license`

ClearKey License Server，供 EME 取得解密金鑰。

- Content-Type: `application/json`
- 請求: `{ "kids": ["<base64url key_id>"], "type": "temporary" }`
- 回應: JWK 格式的金鑰

### `GET /api/videos`

列出所有已加密的影片。

## 切換到 Widevine

ClearKey 與 Widevine 使用相同的 CENC 加密格式。若要切換到正式 Widevine DRM：

1. 取得 [Widevine 合作夥伴](https://www.widevine.com/contact) 憑證
2. 將 `/api/license` 改為 Widevine License Proxy
3. 前端 DRM scheme 從 `org.w3.clearkey` 改為 `com.widevine.alpha`
