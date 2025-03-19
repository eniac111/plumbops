package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"

	"github.com/eniac111/plumbops/internal/types"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Connect opens an SSH connection using user/password or user/key auth.
func Connect(host types.Host) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	if host.Password != "" {
		authMethods = append(authMethods, ssh.Password(host.Password))
	}

	if host.KeyPath != "" {
		key, err := os.ReadFile(host.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSH key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Always try to use the default SSH key if no key path is provided
	if host.KeyPath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("failed to get current user: %w", err)
		}
		defaultKeyPath := filepath.Join(usr.HomeDir, ".ssh", "id_rsa")
		key, err := os.ReadFile(defaultKeyPath)
		if err == nil {
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
				fmt.Println("Using default SSH key:", defaultKeyPath)
			} else {
				fmt.Println("Failed to parse default SSH key:", err)
			}
		} else {
			fmt.Println("Failed to read default SSH key:", err)
		}
	}

	// Always try to use the SSH agent
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		authMethods = append(authMethods, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
		fmt.Println("Using SSH agent")
	} else {
		fmt.Println("Failed to connect to SSH agent:", err)
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}

	config := &ssh.ClientConfig{
		User:            host.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // DO NOT USE IN PRODUCTION
	}

	addr := fmt.Sprintf("%s:%d", host.Name, host.Port)
	if host.Port == 0 {
		addr = fmt.Sprintf("%s:22", host.Name) // default port 22
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}
	return client, nil
}

// UploadFile uses SFTP to copy a local file to a remote path.
func UploadFile(sshClient *ssh.Client, localPath, remotePath string) error {
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// UploadBytes uses SFTP to copy in-memory bytes to a remote file.
func UploadBytes(sshClient *ssh.Client, data []byte, remotePath string) error {
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.Write(data)
	return err
}

// RunCommand executes a command on the remote host via SSH.
func RunCommand(sshClient *ssh.Client, cmd string) (string, string, error) {
	session, err := sshClient.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	// Capture stdout, stderr
	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return "", "", err
	}

	if err := session.Run(cmd); err != nil {
		// read stdout/stderr anyway before returning
		outBytes, _ := io.ReadAll(stdout)
		errBytes, _ := io.ReadAll(stderr)
		return string(outBytes), string(errBytes), err
	}

	outBytes, _ := io.ReadAll(stdout)
	errBytes, _ := io.ReadAll(stderr)
	return string(outBytes), string(errBytes), nil
}
