#!/bin/bash

# FPS2X Go 版本构建脚本

set -e

VERSION=${VERSION:-"1.0.0"}
BUILD_DIR="build"
APP_NAME="FPS2X"

# 设置 SDK 路径（macOS 需要）
SDK_PATH=$(xcrun --sdk macosx --show-sdk-path)
export CGO_CFLAGS="-isysroot $SDK_PATH"
export CGO_LDFLAGS="-isysroot $SDK_PATH"

echo "🚀 开始构建 FPS2X Go 版本..."

# 清理旧的构建
echo "📁 清理旧的构建文件..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# 下载依赖
echo "📦 下载 Go 依赖..."
go mod tidy

# 检测操作系统
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
    Darwin)
        echo "🍎 构建 macOS 版本..."

        # 构建可执行文件
        go build -ldflags="-s -w" -o "$BUILD_DIR/$APP_NAME" main.go

        # 创建 .app 包
        APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"
        mkdir -p "$APP_BUNDLE/Contents/MacOS"
        mkdir -p "$APP_BUNDLE/Contents/Resources"

        # 复制可执行文件
        cp "$BUILD_DIR/$APP_NAME" "$APP_BUNDLE/Contents/MacOS/"

        # 复制 binaries
        cp -r binaries "$APP_BUNDLE/Contents/Resources/"

        # 复制图标
        cp fps2x.icns "$APP_BUNDLE/Contents/Resources/"

        # 创建 Info.plist
        cat > "$APP_BUNDLE/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>$APP_NAME</string>
    <key>CFBundleIconFile</key>
    <string>fps2x.icns</string>
    <key>CFBundleIdentifier</key>
    <string>com.fps2x.desktop</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleSignature</key>
    <string>????</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
</dict>
</plist>
EOF

        echo "✅ macOS .app 包已创建: $APP_BUNDLE"

        # 获取文件大小
        SIZE=$(du -sh "$APP_BUNDLE" | cut -f1)
        echo "📊 应用大小: $SIZE"

        ;;
    Linux)
        echo "🐧 构建 Linux 版本..."

        # 构建可执行文件
        go build -ldflags="-s -w" -o "$BUILD_DIR/$APP_NAME" main.go

        # 创建发布包
        RELEASE_DIR="$BUILD_DIR/$APP_NAME-linux"
        mkdir -p "$RELEASE_DIR"
        cp "$BUILD_DIR/$APP_NAME" "$RELEASE_DIR/"
        cp -r binaries "$RELEASE_DIR/"

        # 打包成 tar.gz
        tar czf "$BUILD_DIR/$APP_NAME-linux-$ARCH.tar.gz" -C "$BUILD_DIR" "$APP_NAME-linux"

        echo "✅ Linux 版本已创建: $BUILD_DIR/$APP_NAME-linux-$ARCH.tar.gz"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        echo "🪟 构建 Windows 版本..."

        # 构建可执行文件
        go build -ldflags="-s -w" -o "$BUILD_DIR/$APP_NAME.exe" main.go

        echo "✅ Windows 版本已创建: $BUILD_DIR/$APP_NAME.exe"
        ;;
    *)
        echo "❌ 不支持的操作系统: $OS"
        exit 1
        ;;
esac

echo ""
echo "🎉 构建完成！"
echo ""
echo "输出目录: $BUILD_DIR"
echo ""
echo "运行应用："
if [ "$OS" = "Darwin" ]; then
    echo "  open $APP_BUNDLE"
elif [ "$OS" = "Linux" ]; then
    echo "  ./$BUILD_DIR/$APP_NAME"
fi
