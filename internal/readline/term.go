package readline

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func GetTerminalMode(file *os.File) (unix.Termios, error) {
	canon := unix.Termios{}
	if _, _, err := unix.Syscall6(unix.SYS_IOCTL,
		file.Fd(), unix.TCGETS, uintptr(unsafe.Pointer(&canon)), 0, 0, 0); err != 0 {
		return canon, err
	}
	return canon, nil
}

func SetTerminalRawMode(file *os.File, canon unix.Termios) (unix.Termios, error) {
	raw := canon
	raw.Iflag &^= (unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP |
		unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON)
	raw.Oflag &^= unix.OPOST
	raw.Lflag &^= (unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN)
	raw.Cflag &^= (unix.CSIZE | unix.PARENB)
	raw.Cflag |= unix.CS8
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0
	if _, _, err := unix.Syscall6(unix.SYS_IOCTL,
		file.Fd(), unix.TCSETS, uintptr(unsafe.Pointer(&raw)), 0, 0, 0); err != 0 {
		return raw, err
	}
	return raw, nil
}

func ResetTerminalRawMode(file *os.File, canon unix.Termios) error {
	if _, _, err := unix.Syscall6(unix.SYS_IOCTL,
		file.Fd(), unix.TCSETS, uintptr(unsafe.Pointer(&canon)), 0, 0, 0); err != 0 {
		return err
	}
	return nil
}
