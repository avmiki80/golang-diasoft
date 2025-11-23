package main

import (
	"errors"
	"io"
	"os"

	"github.com/cheggaaa/pb"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

type FromCopy struct {
	From *os.File
}

type Checker interface {
	CheckFile() error
	CheckOffset(offset int64) error
	CheckLimit(offset, limit int64) (int64, error)
}

func NewFromCopy(fromPath string) (*FromCopy, error) {
	file, err := os.Open(fromPath)
	if err != nil {
		return nil, err
	}
	return &FromCopy{From: file}, nil
}

func (f *FromCopy) Close() error {
	return f.From.Close()
}

func (f *FromCopy) CheckFile() error {
	fromInfo, err := f.From.Stat()
	if err != nil {
		return err
	}
	if fromInfo.IsDir() {
		return ErrUnsupportedFile
	}
	return nil
}

func (f *FromCopy) CheckOffset(offset int64) error {
	fromInfo, err := f.From.Stat()
	if err != nil {
		return err
	}
	if offset > fromInfo.Size() {
		return ErrOffsetExceedsFileSize
	}
	return nil
}

func (f *FromCopy) CheckLimit(offset, limit int64) (int64, error) {
	fromInfo, err := f.From.Stat()
	if err != nil {
		return -1, err
	}
	fileSize := fromInfo.Size()
	if limit < 0 {
		limit = fileSize - offset
	}
	if limit > 0 && offset+limit > fileSize {
		limit = fileSize - offset
	}
	return limit, nil
}

func (f *FromCopy) Seek(offset int64) error {
	_, err := f.From.Seek(offset, io.SeekStart)
	return err
}

type ToCopy struct {
	To *os.File
}

func NewToCopy(toPath string) (*ToCopy, error) {
	file, err := os.Create(toPath)
	if err != nil {
		return nil, err
	}
	return &ToCopy{To: file}, nil
}

func (t *ToCopy) Close() error {
	return t.To.Close()
}

func Copy(fromPath, toPath string, offset, limit int64) error {
	from, err := NewFromCopy(fromPath)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := NewToCopy(toPath)
	if err != nil {
		return err
	}
	defer to.Close()

	err = from.CheckFile()
	if err != nil {
		return err
	}

	err = from.CheckOffset(offset)
	if err != nil {
		return err
	}

	limit, err = from.CheckLimit(offset, limit)
	if err != nil {
		return err
	}
	err = from.Seek(offset)
	if err != nil {
		return err
	}
	err = processCopy(from, to, limit)
	if err != nil {
		return err
	}

	return nil
}

func processCopy(from *FromCopy, to *ToCopy, limit int64) error {
	if limit == 0 {
		// Получаем размер файла для прогресс-бара
		fileInfo, err := from.From.Stat()
		if err != nil {
			return err
		}
		fileSize := fileInfo.Size()

		// Создаем прогресс-бар
		bar := pb.New64(fileSize).SetUnits(pb.U_BYTES)
		bar.Start()

		// Создаем прокси-ридер с прогресс-баром
		proxyReader := bar.NewProxyReader(from.From)

		// Копируем данные с отображением прогресса
		_, err = io.Copy(to.To, proxyReader)
		bar.Finish()
		return err
	}

	// Создаем прогресс-бар для ограниченного копирования
	bar := pb.New64(limit).SetUnits(pb.U_BYTES)
	bar.Start()

	// Создаем прокси-ридер с прогресс-баром
	proxyReader := bar.NewProxyReader(from.From)

	// Копируем ограниченное количество данных с отображением прогресса
	_, err := io.CopyN(to.To, proxyReader, limit)
	bar.Finish()
	return err
}
