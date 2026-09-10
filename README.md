# dandan-prime


[English](#english) | [中文](#中文)
WebUI in ：[dandan-prime-web](https://github.com/DuckFeather10086/dandan-prime-web)  <!-- 添加 dandan-prime-web 链接 -->


![Screenshot](/MainScreen.png)  <!-- 添加截图 -->

![Screenshot](/Player.png)  <!-- 添加截图 -->


## English

dandan-prime is a local streaming media server that supports danmaku (bullet comments) and HLS streaming. It integrates with the Bangumi database for metadata scraping and is written in Go using the Echo framework.

### Features

- Automatic metadata scraping
- Subtitle matching
- HLS streaming

### Dependencies

- Go
- ffmpeg (installed in system path)

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/DuckFeather10086/dandan-prime.git
   ```

2. Navigate to the project directory:
   ```
   cd dandan-prime/cmd
   ```

3. Create a `config.json` file with the following content:
   ```json
   {
       "media_library_root_path": "YOUR_MEDIA_LIBRARY_ROOT_PATH",
       "allowed_video_extensions": [
           ".mp4",
           ".mkv",
           ".avi",
           ".mov",
           ".wmv",
           ".flv",
           ".mpg",
           ".mpeg"
       ],
       "use_hls": true,
       "hls_cache_path": "cache"
   }
   ```
   You can change the options to meet your requirements

4. Run the application:
   ```
   go run ./main.go
   ```

### Acknowledgements

- [Bangumi API](https://github.com/bangumi/api/)
- [dandanplay-libraryindex](https://github.com/kaedei/dandanplay-libraryindex)
- [ffmpeg](https://ffmpeg.org/)  <!-- 添加 ffmpeg 链接 -->

## 中文

dandan-prime 是一个支持弹幕和 HLS 推流的本地流媒体服务器。它集成了 Bangumi 数据库进行元数据刮削，使用 Go 语言和 Echo 框架编写。

### 主要功能和特点

- 自动刮削元数据
- 字幕匹配
- HLS 推流

### 依赖项

- Go
- ffmpeg (安装在系统目录下)

### 安装步骤

1. 克隆仓库：
   ```
   git clone https://github.com/DuckFeather10086/dandan-prime.git
   ```

2. 进入项目目录：
   ```
   cd dandan-prime/cmd
   ```

3. 创建 `config.json` 文件，内容如下：
   ```json
   {
       "media_library_root_path": "你的媒体库根路径",
       "allowed_video_extensions": [
           ".mp4",
           ".mkv",
           ".avi",
           ".mov",
           ".wmv",
           ".flv",
           ".mpg",
           ".mpeg"
       ],
       "use_hls": true,
       "hls_cache_path": "cache"
   }
   ```
   您可以修改配置中的选项来对应需求

4. 运行应用：
   ```
   go run ./main.go
   ```

### 鸣谢

- [Bangumi API](https://github.com/bangumi/api/)
- [dandanplay-libraryindex](https://github.com/kaedei/dandanplay-libraryindex)
- [ffmpeg](https://ffmpeg.org/)  <!-- 添加 ffmpeg 链接 -->
---

## Matching against bangumi.tv directly

Metadata used to be reached through dandanplay: a file hash went to its
`/match` endpoint for an anime id, and that anime's `OnlineDatabases` link gave
the bangumi.tv subject id the metadata actually came from. That made
dandanplay's granularity the library's granularity, which breaks on film
series — all seven *Kara no Kyoukai* films share one dandanplay anime id, while
bangumi.tv models them as seven separate subjects. The hardcoded Evangelion id
fixes in `usecase/bangumiUseCase` were the same failure, patched one id at a
time. A file whose hash is absent from dandanplay's index also never matched,
and no amount of rescanning helped.

`usecase/matchUseCase` searches bangumi.tv directly instead, and asks an
OpenAI-compatible model to choose among the candidates it returns. The model is
used for one judgement only — which candidate a directory refers to, including
across languages, where character similarity is useless (`Dungeon Meshi` vs
`ダンジョン飯` overlaps at ~0, yet bangumi.tv's own search ranks it first). It is
never a metadata source: any subject id it returns that was not in the
candidate list is rejected before anything is written.

Resolution happens per top-level directory but is assigned per file, so a
directory holding a film series maps each film to its own subject.

Hashes are still computed at scan time and still used for danmaku, which is the
one thing dandanplay is authoritative about.

### Configuration

`config.json` gains two optional fields:

```json
{
    "llm_base_url": "http://127.0.0.1:8650/v1",
    "llm_model": "gemini-3.8-flash"
}
```

The API key is read from the environment, never from `config.json`, so it stays
out of this repository:

| Variable | Purpose |
|---|---|
| `DANDAN_LLM_API_KEY` | Bearer token for the endpoint. Required. |
| `DANDAN_LLM_BASE_URL` | Overrides `llm_base_url`. |
| `DANDAN_LLM_MODEL` | Overrides `llm_model`. |

Any OpenAI-compatible endpoint works. Under systemd, keep them in a 0600 file
loaded with `EnvironmentFile=-%h/.config/dandan-prime/env`.

With no key configured the scan still runs; directories are simply left
unresolved rather than failing.

### Usage

`POST /api/bangumi/media-library` runs the whole chain. Each stage is also its
own endpoint, because they have very different costs and you rarely want all
of them:

| Endpoint | Cost | Does |
|---|---|---|
| `media-library` | | The four below, in order |
| `episode-numbers` | free | Re-derives episode numbers from file names |
| `resolve` | LLM | Matches directories to bangumi.tv subjects |
| `episode-metadata` | bangumi.tv only | Per-episode titles and summaries |

All take `?force=true` to redo work already done; `resolve` also takes
`?limit=N` to work through a large library in batches. `resolve` returns
counts of what it did, including the tokens spent.

`episode-numbers` exists so that improving the file-name parser does not mean
re-resolving the library and paying for the LLM calls again — episode numbers
come from the names, not from the model, so they can be recomputed offline.

Danmaku still needs dandanplay, and `media-library` asks it for episode ids
when `dandan_play_app_id` and `dandan_play_app_secret` are set. Nothing else
depends on it: without credentials the library builds normally and only
danmaku is missing.

### Fixing a bad match

Put a `.bangumi-id` file in the directory. It is honoured before the model is
consulted, so a correction survives every later rescan:

```
# whole directory is one subject
1671
```

```
# one line per file, for a film collection
[SumiSora][Kara_no_Kyoukai][01].mkv	233
[SumiSora][Kara_no_Kyoukai][02].mkv	812
```

File names are relative to the directory holding the `.bangumi-id`.

### Filename parsing

`usecase/matchUseCase/parse.go` is deliberately shallow. Rule-based title
extraction was tried against a 189-directory library and leaked in both
directions: picking the longest bracket block yields the release group name
(`Nekomoe kissaten&VCB-Studio`), and filtering technical tags yields the
leftovers (`HEVC-YUV420P10`). So the keyword it produces is only used to fetch
candidates — a bad keyword costs one extra round trip, never a wrong match,
because the model is shown the raw directory name. Episode *numbers* stay
rule-based, since they are unambiguous once the technical tokens are out of the
way.

To inspect what the parser cannot read, against your own library:

```sh
sqlite3 cmd/media_library.db 'select file_name from episode_infos' > /tmp/corpus.txt
DANDAN_PARSE_CORPUS=/tmp/corpus.txt go test ./usecase/matchUseCase/ -run Corpus -v
```
