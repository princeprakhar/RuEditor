package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rueditor/backend/internal/models"
)

type LatexService struct {
	workDir string
}

type CompileResult struct {
	PDF []byte
	Log string
}

func NewLatexService() *LatexService {
	workDir := "/app/data"
	os.MkdirAll(filepath.Join(workDir, "documents"), 0755)
	os.MkdirAll(filepath.Join(workDir, "compiled"), 0755)
	os.MkdirAll(filepath.Join(workDir, "temp"), 0755)

	return &LatexService{
		workDir: workDir,
	}
}

func (s *LatexService) Compile(content string, docID string) (*CompileResult, error) {
	docDir := filepath.Join(s.workDir, "documents", docID)
	tempDir := filepath.Join(s.workDir, "temp", docID+"-"+time.Now().Format("20060102150405"))
	compiledDir := filepath.Join(s.workDir, "compiled", docID)

	os.MkdirAll(docDir, 0755)
	os.MkdirAll(tempDir, 0755)
	os.MkdirAll(compiledDir, 0755)

	mainTexPath := filepath.Join(tempDir, "main.tex")
	if err := os.WriteFile(mainTexPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write LaTeX file: %w", err)
	}

	logBuilder := &strings.Builder{}
	var stderr bytes.Buffer

	cmd := exec.Command("pdflatex", "-interaction=nonstopmode", "-halt-on-error",
		"-output-directory="+tempDir, "main.tex")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(),
		"TEXMFHOME="+filepath.Join(s.workDir, "texmf"),
		"TEXMFCONFIG="+filepath.Join(s.workDir, "texmf-config"),
		"TEXMFVAR="+filepath.Join(s.workDir, "texmf-var"),
	)
	cmd.Stdout = logBuilder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		logBuilder.WriteString("\n[ERROR] ")
		logBuilder.WriteString(stderr.String())
		return &CompileResult{
			Log: logBuilder.String(),
		}, fmt.Errorf("LaTeX compilation failed: %w", err)
	}

	pdfPath := filepath.Join(tempDir, "main.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return &CompileResult{
			Log: logBuilder.String(),
		}, fmt.Errorf("failed to read PDF: %w", err)
	}

	os.WriteFile(filepath.Join(compiledDir, "main.pdf"), pdfBytes, 0644)

	s.cleanupOldFiles(tempDir, 24*time.Hour)

	return &CompileResult{
		PDF: pdfBytes,
		Log: logBuilder.String(),
	}, nil
}

func (s *LatexService) CompileWithBiber(content string, docID string) (*CompileResult, error) {
	docDir := filepath.Join(s.workDir, "documents", docID)
	tempDir := filepath.Join(s.workDir, "temp", docID+"-"+time.Now().Format("20060102150405"))
	compiledDir := filepath.Join(s.workDir, "compiled", docID)

	os.MkdirAll(docDir, 0755)
	os.MkdirAll(tempDir, 0755)
	os.MkdirAll(compiledDir, 0755)

	mainTexPath := filepath.Join(tempDir, "main.tex")
	if err := os.WriteFile(mainTexPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write LaTeX file: %w", err)
	}

	logBuilder := &strings.Builder{}
	var stderr bytes.Buffer

	commands := []struct {
		name string
		cmd  *exec.Cmd
	}{
		{
			name: "pdflatex",
			cmd: exec.Command("pdflatex", "-interaction=nonstopmode", "-halt-on-error",
				"-output-directory="+tempDir, "main.tex"),
		},
		{
			name: "biber",
			cmd: exec.Command("biber", "--input-directory="+tempDir,
				"--output-directory="+tempDir, "main"),
		},
		{
			name: "pdflatex",
			cmd: exec.Command("pdflatex", "-interaction=nonstopmode", "-halt-on-error",
				"-output-directory="+tempDir, "main.tex"),
		},
		{
			name: "pdflatex",
			cmd: exec.Command("pdflatex", "-interaction=nonstopmode", "-halt-on-error",
				"-output-directory="+tempDir, "main.tex"),
		},
	}

	for i, c := range commands {
		c.cmd.Dir = tempDir
		c.cmd.Stdout = logBuilder
		c.cmd.Stderr = &stderr

		if err := c.cmd.Run(); err != nil {
			logBuilder.WriteString(fmt.Sprintf("\n[%s (attempt %d) ERROR] %s\n", c.name, i+1, stderr.String()))
			if i < 2 {
				stderr.Reset()
				continue
			}
			return &CompileResult{
				Log: logBuilder.String(),
			}, fmt.Errorf("%s compilation failed: %w", c.name, err)
		}
		stderr.Reset()
	}

	pdfPath := filepath.Join(tempDir, "main.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return &CompileResult{
			Log: logBuilder.String(),
		}, fmt.Errorf("failed to read PDF: %w", err)
	}

	os.WriteFile(filepath.Join(compiledDir, "main.pdf"), pdfBytes, 0644)

	s.cleanupOldFiles(tempDir, 24*time.Hour)

	return &CompileResult{
		PDF: pdfBytes,
		Log: logBuilder.String(),
	}, nil
}

func (s *LatexService) GetCompiledPDF(docID string) ([]byte, error) {
	pdfPath := filepath.Join(s.workDir, "compiled", docID, "main.pdf")
	return os.ReadFile(pdfPath)
}

func (s *LatexService) cleanupOldFiles(tempDir string, maxAge time.Duration) {
	files, err := os.ReadDir(filepath.Dir(tempDir))
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-maxAge)
	for _, f := range files {
		if f.IsDir() {
			info, _ := f.Info()
			if info.ModTime().Before(cutoff) {
				os.RemoveAll(filepath.Join(filepath.Dir(tempDir), f.Name()))
			}
		}
	}
}

func (s *LatexService) GetLog(docID string) (*models.CompileResponse, error) {
	return &models.CompileResponse{
		Success: true,
	}, nil
}
