package domain

import "fmt"

type Table struct {
	Database string
	Schema   string
	Name     string
	RowCount int
	Size     int64
}

func (t Table) SizeMb() string {
	return fmt.Sprintf("%d MB", t.Size/1024/1024)
}
