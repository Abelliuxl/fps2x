package main

import (
	"embed"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

//go:embed binaries/*
var binaries embed.FS

var (
	selectedFilePath string
	progressBar      *widget.ProgressBar
	progressLabel    *widget.Label
	statusLabel      *widget.Label
	processBtn       *widget.Button
	selectBtn        *widget.Button
	fileLabel        *widget.Label
	resultLabel      *widget.Label
	mainWindow       fyne.Window

	// 文件卡片元素
	fileCardContainer *fyne.Container
	fileNameLabel     *widget.Label
	filePathLabel     *widget.Label
	fileIconCanvas    *canvas.Text

	// 步骤标签
	stepExtractLabel *widget.Label
	stepInterpLabel  *widget.Label
	stepMergeLabel   *widget.Label
)

type ProcessingStep int

const (
	StepPending ProcessingStep = iota
	StepRunning
	StepCompleted
	StepError
)

type DependencyCheck struct {
	Ready  bool
	Paths  *BinaryPaths
	Error  string
}

type BinaryPaths struct {
	FFmpeg  string
	FFprobe string
	RIFE    string
	Model   string
}

func main() {
	myApp := app.NewWithID("com.fps2x.desktop")

	mainWindow = myApp.NewWindow("FPS2X - 视频帧率倍增器")
	mainWindow.Resize(fyne.NewSize(600, 620))
	mainWindow.CenterOnScreen()

	// 创建 UI
	ui := createUI()
	mainWindow.SetContent(ui)

	// 启动时检查依赖
	go checkDependenciesOnStart()

	mainWindow.ShowAndRun()
}

func createUI() *fyne.Container {
	// 标题
	title := widget.NewLabel("FPS2X")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := widget.NewLabel("视频帧率倍增器")
	subtitle.TextStyle = fyne.TextStyle{Italic: true}
	subtitle.Alignment = fyne.TextAlignCenter

	// 文件状态卡片（大图标显示 - 4倍大小）
	largeFontSize := float32(48) // 约4倍正常字体大小
	fileIconCanvas = canvas.NewText("❓", color.White)
	fileIconCanvas.TextSize = largeFontSize
	fileIconCanvas.Alignment = fyne.TextAlignCenter

	// 用container居中显示图标
	fileIconCentered := container.NewCenter(fileIconCanvas)

	fileNameLabel = widget.NewLabel("未选择文件")
	fileNameLabel.Alignment = fyne.TextAlignCenter
	fileNameLabel.TextStyle = fyne.TextStyle{Bold: true}

	filePathLabel = widget.NewLabel("请选择要处理的视频文件")
	filePathLabel.Alignment = fyne.TextAlignCenter
	filePathLabel.Wrapping = fyne.TextWrapWord

	fileCardContainer = container.NewVBox(
		fileIconCentered,
		container.NewPadded(fileNameLabel),
		container.NewPadded(filePathLabel),
	)

	// 提示标签（保留但初始隐藏，在未选择文件时显示卡片）
	fileLabel = widget.NewLabel("点击下方按钮选择视频文件")
	fileLabel.Alignment = fyne.TextAlignCenter
	fileLabel.Wrapping = fyne.TextWrapWord
	fileLabel.Hide() // 现在始终显示卡片

	// 按钮区域 - 横向布局但不占满宽度
	selectBtn = widget.NewButton("选择视频文件", onSelectFile)
	processBtn = widget.NewButton("开始处理", onProcessVideo)
	processBtn.Disable()

	buttonBox := container.NewHBox(
		selectBtn,
		processBtn,
	)
	// 居中对齐按钮
	buttonBoxCentered := container.NewCenter(buttonBox)

	// 进度区域
	progressLabel = widget.NewLabel("准备就绪")
	progressLabel.TextStyle = fyne.TextStyle{Bold: true}
	progressBar = widget.NewProgressBar()
	progressBar.SetValue(0)

	// 处理步骤（使用更紧凑的布局）
	stepTitle := widget.NewLabel("处理流程")
	stepTitle.TextStyle = fyne.TextStyle{Bold: true}

	stepExtractLabel = createStepLabel("⏳", "提取视频帧")
	stepInterpLabel = createStepLabel("⏳", "AI 插帧")
	stepMergeLabel = createStepLabel("⏳", "合并视频")

	stepsBox := container.NewVBox(
		stepExtractLabel,
		widget.NewSeparator(),
		stepInterpLabel,
		widget.NewSeparator(),
		stepMergeLabel,
	)

	// 状态和结果
	statusLabel = widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord
	statusLabel.Alignment = fyne.TextAlignCenter

	resultLabel = widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord
	resultLabel.Alignment = fyne.TextAlignCenter
	resultLabel.TextStyle = fyne.TextStyle{Bold: true}
	resultLabel.Hide()

	// 主布局
	content := container.NewVBox(
		// 标题区域
		container.NewPadded(title),
		container.NewPadded(subtitle),
		widget.NewSeparator(),

		// 文件选择区域 - 始终显示卡片
		container.NewPadded(fileCardContainer),
		container.NewPadded(buttonBoxCentered),
		widget.NewSeparator(),

		// 进度区域
		container.NewPadded(progressLabel),
		container.NewPadded(progressBar),
		widget.NewSeparator(),

		// 步骤区域
		container.NewPadded(stepTitle),
		container.NewPadded(stepsBox),
		widget.NewSeparator(),

		// 状态区域
		statusLabel,
		resultLabel,
	)

	return container.NewPadded(content)
}

// 创建步骤标签，统一样式
func createStepLabel(icon, text string) *widget.Label {
	label := widget.NewLabel(fmt.Sprintf("%s  %s", icon, text))
	label.TextStyle = fyne.TextStyle{Bold: true}
	return label
}

func onSelectFile() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		// 获取文件路径
		uri := reader.URI()
		selectedFilePath = uri.Path()
		if selectedFilePath == "" {
			// 尝试从 URI 解析
			selectedFilePath = fmt.Sprintf("%s", uri)
		}
		filename := filepath.Base(selectedFilePath)

		// 更新 UI - 显示绿色勾号（4倍大小）
		fileIconCanvas.Text = "✅"
		fileIconCanvas.Color = color.RGBA{0, 200, 0, 255} // 绿色
		fileIconCanvas.Refresh()

		fileNameLabel.SetText(filename)
		filePathLabel.SetText(selectedFilePath)

		processBtn.Enable()
	}, mainWindow)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv"}))
	fd.Show()
}

func onProcessVideo() {
	if selectedFilePath == "" {
		return
	}

	// 禁用按钮
	selectBtn.Disable()
	processBtn.Disable()
	resultLabel.Hide()

	// 重置进度和步骤
	progressBar.SetValue(0)
	statusLabel.SetText("开始处理...")
	resetSteps()

	// 在后台处理视频
	go processVideo(selectedFilePath)
}

func checkDependenciesOnStart() {
	statusLabel.SetText("正在检查依赖...")

	depCheck, err := checkDependencies()
	if err != nil {
		statusLabel.SetText(fmt.Sprintf("依赖检查失败: %v", err))
		return
	}

	if !depCheck.Ready {
		statusLabel.SetText(fmt.Sprintf("依赖错误: %s\n请确保 binaries 目录包含所有必需文件", depCheck.Error))
		dialog.ShowError(fmt.Errorf("依赖检查失败: %s", depCheck.Error), mainWindow)
	} else {
		statusLabel.SetText("依赖检查完成，准备就绪")
	}
}

func checkDependencies() (*DependencyCheck, error) {
	binariesPath, err := getBinariesPath()
	if err != nil {
		return nil, err
	}

	ffmpegPath := filepath.Join(binariesPath, "ffmpeg")
	ffprobePath := filepath.Join(binariesPath, "ffprobe")
	rifePath := filepath.Join(binariesPath, "rife-ncnn-vulkan")
	modelPath := filepath.Join(binariesPath, "rife-v4.6")

	// Windows 添加 .exe 扩展名
	if runtime.GOOS == "windows" {
		ffmpegPath += ".exe"
		ffprobePath += ".exe"
		rifePath += ".exe"
	}

	// 检查文件是否存在
	if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
		return &DependencyCheck{Ready: false, Error: "FFmpeg 未找到"}, nil
	}
	if _, err := os.Stat(ffprobePath); os.IsNotExist(err) {
		return &DependencyCheck{Ready: false, Error: "FFprobe 未找到"}, nil
	}
	if _, err := os.Stat(rifePath); os.IsNotExist(err) {
		return &DependencyCheck{Ready: false, Error: "RIFE 主程序未找到"}, nil
	}
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return &DependencyCheck{Ready: false, Error: "RIFE 模型文件未找到"}, nil
	}

	return &DependencyCheck{
		Ready: true,
		Paths: &BinaryPaths{
			FFmpeg:  ffmpegPath,
			FFprobe: ffprobePath,
			RIFE:    rifePath,
			Model:   modelPath,
		},
	}, nil
}

func getBinariesPath() (string, error) {
	// 开发环境：使用项目根目录的 binaries
	if _, err := os.Stat("binaries"); err == nil {
		return "binaries", nil
	}

	// 生产环境：使用可执行文件所在目录
	if runtime.GOOS == "darwin" {
		exePath, err := os.Executable()
		if err != nil {
			return "", err
		}
		return filepath.Join(filepath.Dir(exePath), "..", "Resources", "binaries"), nil
	}

	// 其他平台：使用可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exePath), "binaries"), nil
}

func processVideo(inputPath string) {
	defer func() {
		fyne.Do(func() {
			selectBtn.Enable()
			processBtn.Enable()
		})
	}()

	// 检查依赖
	depCheck, err := checkDependencies()
	if err != nil {
		showError(fmt.Sprintf("依赖检查失败: %v", err))
		return
	}

	if !depCheck.Ready {
		showError(depCheck.Error)
		return
	}

	paths := depCheck.Paths

	// 创建工作目录
	downloadsPath, err := os.UserHomeDir()
	if err != nil {
		showError(fmt.Sprintf("无法获取用户目录: %v", err))
		return
	}

	downloadsPath = filepath.Join(downloadsPath, "Downloads")
	fileName := filepath.Base(inputPath)
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	workDir := filepath.Join(downloadsPath, fmt.Sprintf("work_%s_%d", baseName, getCurrentTimestamp()))

	// 创建工作目录
	if err := os.MkdirAll(filepath.Join(workDir, "in"), 0755); err != nil {
		showError(fmt.Sprintf("创建工作目录失败: %v", err))
		return
	}
	if err := os.MkdirAll(filepath.Join(workDir, "out"), 0755); err != nil {
		showError(fmt.Sprintf("创建工作目录失败: %v", err))
		return
	}
	defer os.RemoveAll(workDir) // 清理临时文件

	// 1. 获取原始帧率
	updateProgress("正在获取视频信息...", 10)
	fpsOrigin, err := getFrameRate(inputPath, paths.FFprobe)
	if err != nil {
		showError(fmt.Sprintf("获取视频帧率失败: %v", err))
		return
	}
	fpsTarget := fpsOrigin * 2

	updateProgress(fmt.Sprintf("帧率转换: %.0f -> %.0f", fpsOrigin, fpsTarget), 20)

	// 2. 提取音频
	updateProgress("正在提取音频...", 30)
	audioPath := filepath.Join(workDir, "audio.m4a")
	if err := runCommand(paths.FFmpeg, []string{
		"-y", "-i", inputPath, "-vn", "-c:a", "copy", audioPath,
	}); err != nil {
		showError(fmt.Sprintf("提取音频失败: %v", err))
		return
	}

	// 3. 拆帧
	updateStep(stepExtractLabel, StepRunning, "提取视频帧")
	updateProgress("正在拆帧...", 40)
	inputFrames := filepath.Join(workDir, "in", "%08d.jpg")
	if err := runCommand(paths.FFmpeg, []string{
		"-y", "-i", inputPath, "-q:v", "2", inputFrames,
	}); err != nil {
		updateStep(stepExtractLabel, StepError, "提取视频帧")
		showError(fmt.Sprintf("拆帧失败: %v", err))
		return
	}
	updateStep(stepExtractLabel, StepCompleted, "提取视频帧")

	// 4. RIFE 插帧
	updateStep(stepInterpLabel, StepRunning, "AI 插帧")
	updateProgress("AI 插帧中（这可能需要几分钟）...", 60)
	if err := runCommand(paths.RIFE, []string{
		"-i", filepath.Join(workDir, "in"),
		"-o", filepath.Join(workDir, "out"),
		"-j", "2:2:2",
		"-m", paths.Model,
	}); err != nil {
		updateStep(stepInterpLabel, StepError, "AI 插帧")
		showError(fmt.Sprintf("AI 插帧失败: %v", err))
		return
	}
	updateStep(stepInterpLabel, StepCompleted, "AI 插帧")

	// 5. 合并视频
	updateStep(stepMergeLabel, StepRunning, "合并视频")
	updateProgress("正在封装最终视频...", 80)
	outputPath := filepath.Join(downloadsPath, fmt.Sprintf("%s_%.0ffps.mp4", baseName, fpsTarget))

	// 根据平台选择编码器
	codec := "libx264"
	if runtime.GOOS == "darwin" {
		codec = "h264_videotoolbox"
	}

	if err := runCommand(paths.FFmpeg, []string{
		"-y", "-framerate", fmt.Sprintf("%.0f", fpsTarget),
		"-i", filepath.Join(workDir, "out", "%08d.png"),
		"-i", audioPath,
		"-c:v", codec,
		"-b:v", "15M",
		"-pix_fmt", "yuv420p",
		"-c:a", "copy",
		"-shortest", outputPath,
	}); err != nil {
		updateStep(stepMergeLabel, StepError, "合并视频")
		showError(fmt.Sprintf("封装视频失败: %v", err))
		return
	}
	updateStep(stepMergeLabel, StepCompleted, "合并视频")

	// 完成
	updateProgress("处理完成！", 100)
	fyne.Do(func() {
		resultLabel.SetText(fmt.Sprintf("视频已保存至:\n%s", outputPath))
		resultLabel.Show()
		dialog.ShowInformation("处理完成", fmt.Sprintf("视频已保存至:\n%s", outputPath), mainWindow)
	})
}

func getFrameRate(inputPath, ffprobePath string) (float64, error) {
	cmd := exec.Command(ffprobePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=r_frame_rate",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("执行 ffprobe 失败: %w", err)
	}

	fpsStr := strings.TrimSpace(string(output))
	parts := strings.Split(fpsStr, "/")
	if len(parts) == 2 {
		numerator := parseFloat(parts[0])
		denominator := parseFloat(parts[1])
		if denominator != 0 {
			return numerator / denominator, nil
		}
	}

	return parseFloat(fpsStr), nil
}

func runCommand(command string, args []string) error {
	log.Printf("Running: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("命令执行失败: %w\n输出: %s", err, string(output))
	}

	return nil
}

func updateProgress(text string, progress float64) {
	fyne.Do(func() {
		progressLabel.SetText(text)
		progressBar.SetValue(progress / 100)
	})
}

func updateStep(stepLabel *widget.Label, status ProcessingStep, stepName string) {
	var icon string
	var text string

	switch status {
	case StepPending:
		icon = "⏳"
		text = stepName
	case StepRunning:
		icon = "🔄"
		text = fmt.Sprintf("正在%s...", stepName)
	case StepCompleted:
		icon = "✅"
		text = fmt.Sprintf("%s完成", stepName)
	case StepError:
		icon = "❌"
		text = fmt.Sprintf("%s失败", stepName)
	}

	fyne.Do(func() {
		stepLabel.SetText(fmt.Sprintf("%s %s", icon, text))
	})
}

func resetSteps() {
	stepExtractLabel.SetText("⏳  提取视频帧")
	stepInterpLabel.SetText("⏳  AI 插帧")
	stepMergeLabel.SetText("⏳  合并视频")
}

func showError(message string) {
	fyne.Do(func() {
		statusLabel.SetText(fmt.Sprintf("错误: %s", message))
		dialog.ShowError(fmt.Errorf("%s", message), mainWindow)
	})
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
