package main

import "os"

type WrapperResp struct {
	Data interface{} `json:"data"`
	Code int         `json:"code"` // 0 - ok, 1 - error
	Msg  string      `json:"msg"`
}

type GeneralizedReport struct {
	Name              string `json:"name"`
	IsDir             bool   `json:"isDir"`
	IsFile            bool   `json:"isFile"`
	IsRead            bool   `json:"isRead"`
	Ext               string `json:"ext"`
	IsPdf             bool   `json:"isPdf"`
	FileCnt           int    `json:"fileCnt"`
	PdfCnt            int    `json:"pdfCnt"`
	PdfReadCnt        int    `json:"pdfReadCnt"`
	SelfPath          string `json:"selfPath"`
	Level             int    `json:"level"`
	ParentPath        string `json:"parentPath"`
	IsIndustryFolder  bool   `json:"isIndustryFolder"`
	HasUnreadIndustry bool   `json:"hasUnreadIndustry"`
}

type DirVal struct {
	Path  string
	Entry os.DirEntry
}

type MultiWayDirNodeV1 struct {
	Val      *DirVal
	Children []*MultiWayDirNodeV1
	Visited  int
}

type MultiWayDirNode struct {
	Val        *DirVal
	Children   []*MultiWayDirNode
	Visited    int
	ParentPath string
}
