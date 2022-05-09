package vscode

import (
	"aliyun/serverless/webide-server/pkg/context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const gOssBucketName string = "hryang-serverless"

func TestLoadEmptyWorkspace(t *testing.T) {
	ctx, err := context.NewFromEnvVars()
	if err != nil {
		t.Fatalf("%v", err)
	}

	vserver := &Server{}
	vserver.OssBucketName = "hryang-serverless"
	ossEndpoint := "https://oss-" + ctx.Region + ".aliyuncs.com"
	c, err := oss.New(ossEndpoint, ctx.AccessKeyId, ctx.AccessKeySecret, oss.SecurityToken(ctx.SecurityToken))
	if err != nil {
		t.Fatalf("Create oss client failed. Context: %v Error: %v", ctx, err)
	}
	vserver.OssClient = c

	// Load workspace from oss and extract data to local directory.
	vserver.WorkspaceDir, err = os.MkdirTemp("", "")
	defer os.RemoveAll(vserver.WorkspaceDir)
	if err != nil {
		t.Fatalf("unable to create temporary dir: %s", vserver.WorkspaceDir)
	}
	vserver.WorkspaceOssPath = "dummy"
	if err = vserver.load(vserver.WorkspaceOssPath, vserver.WorkspaceDir); err != nil {
		t.Fatalf("unable to load workspace. error: %v", err)
	}

	// Expect the local directory is empty.
	dir, _ := ioutil.ReadDir(vserver.WorkspaceDir)
	if len(dir) != 0 {
		t.Fatalf("expect empty directory, but got %v", dir)
	}
}

func TestWorkspace(t *testing.T) {
	ctx, err := context.NewFromEnvVars()
	if err != nil {
		t.Fatalf("%v", err)
	}

	vserver := &Server{}
	vserver.OssBucketName = "hryang-serverless"
	vserver.WorkspaceOssPath = "tests/vscode-server/test-workspace/workspace.tar.gz"
	ossEndpoint := "https://oss-" + ctx.Region + ".aliyuncs.com"
	c, err := oss.New(ossEndpoint, ctx.AccessKeyId, ctx.AccessKeySecret, oss.SecurityToken(ctx.SecurityToken))
	if err != nil {
		t.Fatalf("Create oss client failed. Context: %v Error: %v", ctx, err)
	}
	vserver.OssClient = c

	srcTemp, err := os.MkdirTemp("", "")
	defer os.RemoveAll(srcTemp)
	if err != nil {
		t.Fatalf("unable to create temporary dir: %s", srcTemp)
	}

	// Prepare the mock workspace data.
	vserver.WorkspaceDir = srcTemp
	filePath := filepath.Join(vserver.WorkspaceDir, "file1.txt")
	cmd := exec.Command("bash", "-c", "mkdir -p "+vserver.WorkspaceDir+"&&"+" echo \"this is file1.\" >> "+filePath)
	err = cmd.Run()
	if err != nil {
		t.Fatalf("unable to create file %s. Error: %v", filePath, err)
	}
	subDir := filepath.Join(vserver.WorkspaceDir, "file2")
	filePath = filepath.Join(subDir, "file2.txt")
	cmd = exec.Command("bash", "-c", "mkdir -p "+subDir+"&&"+" echo \"this is file2.\" >> "+filePath)
	err = cmd.Run()
	if err != nil {
		t.Fatalf("unable to create file %s. Error: %v", filePath, err)
	}

	// Save workspace data to oss.
	if err = vserver.save(vserver.WorkspaceDir, vserver.WorkspaceOssPath); err != nil {
		t.Fatalf("unable to save workspace. error: %v", err)
	}

	// Load workspace data from oss.
	dstTemp, err := os.MkdirTemp("", "")
	defer os.RemoveAll(dstTemp)
	if err != nil {
		t.Fatalf("unable to create temporary dir: %s", dstTemp)
	}
	vserver.WorkspaceDir = dstTemp
	if err = vserver.load(vserver.WorkspaceOssPath, vserver.WorkspaceDir); err != nil {
		t.Fatalf("unable to load workspace. error: %v", err)
	}

	// Verify the data.
	cmd = exec.Command("diff", "--recursive", srcTemp, dstTemp)
	err = cmd.Run()
	if err != nil {
		t.Fatalf("The two directories are not equal.\nSrc dir: %s\nDst dir: %s\nError: %v", srcTemp, dstTemp, err)
	}
}
