# FPS2X - 视频帧率倍增器

基于 Go + Fyne 的视频帧率倍增工具，使用 RIFE AI 模型进行视频插帧。

## 特性

- 🚀 **轻量级**：相比 Electron 版本减少约 70% 体积
- ⚡ **原生性能**：Go 原生应用，启动速度快，内存占用低
- 🎨 **简洁 UI**：基于 Fyne 的跨平台界面
- 🤖 **AI 插帧**：使用 RIFE 模型将视频帧率倍增（30fps → 60fps）

## 体积对比

| 版本 | 体积 |
|------|------|
| Electron 版 | ~100MB |
| Go 版本 | ~35MB |

## 依赖项

- Go 1.21+
- FFmpeg 4.0+
- FFprobe
- RIFE-NCNN-Vulkan
- RIFE v4.6 模型

## 快速开始

### 开发模式运行

```bash
go run main.go
```

### 构建应用

使用提供的构建脚本：

```bash
./build.sh
```

这会自动创建对应平台的分发包。

## 手动构建

### macOS

```bash
# 下载依赖
go mod tidy

# 构建可执行文件
go build -o fps2x main.go

# 运行
./fps2x
```

### Linux

```bash
GOOS=linux GOARCH=amd64 go build -o fps2x main.go
```

### Windows

```bash
GOOS=windows GOARCH=amd64 go build -o fps2x.exe main.go
```

## 打包说明

`build.sh` 脚本会自动处理打包流程：

- **macOS**: 创建 `.app` 包，包含所有二进制依赖
- **Linux**: 创建 tar.gz 压缩包
- **Windows**: 创建包含所有文件的发布包

## 使用方法

1. 启动应用
2. 点击"选择视频文件"按钮
3. 选择要处理的视频文件
4. 点击"开始处理"
5. 等待处理完成（可能需要几分钟，取决于视频长度）
6. 输出文件保存在 `~/Downloads/` 文件夹

## 支持的视频格式

- MP4
- AVI
- MOV
- MKV
- WMV
- FLV

## 项目结构

```
.
├── main.go          # 主程序和 UI
├── go.mod           # Go 模块文件
├── go.sum           # 依赖锁定
├── build.sh         # 构建脚本
├── binaries/        # 二进制依赖
│   ├── ffmpeg
│   ├── ffprobe
│   ├── rife-ncnn-vulkan
│   └── rife-v4.6/   # RIFE 模型文件
└── README.md
```

## 技术栈

- **UI 框架**: Fyne 2.4+
- **编程语言**: Go 1.21+
- **视频处理**: FFmpeg
- **AI 插帧**: RIFE-NCNN-Vulkan

## 工作流程

1. **提取音频**: 使用 FFmpeg 从原视频提取音频
2. **拆帧**: 将视频拆分为独立帧（JPG 格式）
3. **AI 插帧**: 使用 RIFE 模型在帧之间插值，生成中间帧
4. **封装**: 将插帧后的图片和音频封装为最终视频

## 系统要求

- **macOS**: 10.15+
- **Linux**: 主流发行版（需要 Vulkan 支持）
- **Windows**: 10+

## 许可证

MIT License

## 致谢

- [Fyne](https://fyne.io/) - 跨平台 UI 框架
- [RIFE](https://github.com/megvii-research/ECCV2022-RIFE) - AI 插帧算法
- [FFmpeg](https://ffmpeg.org/) - 视频处理工具
- [ncnn](https://github.com/Tencent/ncnn) - 高性能神经网络推理框架
