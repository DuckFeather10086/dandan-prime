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