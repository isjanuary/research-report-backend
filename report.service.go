package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

type ReportService struct{}

// const defaultDir = "C:\\Users\\Jiayi Xu\\Desktop\\重要文件\\研报"

func NewReportService() *ReportService {
	return &ReportService{}
}

func (rSvc *ReportService) InitReadStatusOfAllReportsV1(rootPath string, ctx *gin.Context) ([]*GeneralizedReport, error) {
	// step 1: 	遍历 rootPath 构建目录结构树
	rootFolder, err := genDirTreeBySys(rootPath)
	if err != nil {
		return nil, err
	}

	// step 2: 	后序遍历目录结构树计算文件数目/已读数目
	//			输入目录树的根节点, 返回 list, list-item 应包含各目录节点信息, 以及所有子 pdf 数, 所有子 pdf 的已读/已阅数目
	//			前端需要的字段有:	1. key - pathname 	2. childrenKeys		3. readPdfs		4. totalPdfs
	stack := []*MultiWayDirNodeV1{rootFolder}
	var stackTop *MultiWayDirNodeV1

	parentPath := rootPath
	currPath := ""
	reportMap := make(map[string]*GeneralizedReport, 0)
	reportMap[parentPath] = &GeneralizedReport{
		Name:       parentPath,
		IsDir:      true,
		IsFile:     false,
		IsRead:     false,
		Ext:        "",
		IsPdf:      false,
		PdfReadCnt: 0,
		FileCnt:    0,
		PdfCnt:     0,
		SelfPath:   parentPath,
		ParentPath: "",
		Level:      0,
		// HasIndustrialReports: false,
	}

	for len(stack) > 0 {
		stackTop = stack[len(stack)-1]
		parentPath = stackTop.Val.Path

		if len(stackTop.Children) == 0 || stackTop.Visited == len(stackTop.Children)-1 {
			//	1. 	叶子节点场景
			//	或
			// 	2.	其所有子树均已处理的场景
			topEntry := stackTop.Val.Entry
			if topEntry == nil {
				// 根节点的情形, Entry == nil
				// 此时要把所有子孙节点的文件计数加上来
				report4Entry := reportMap[parentPath]
				childrenEntries := stackTop.Children
				fileCnt := 0
				pdfCnt := 0
				pdfReadCnt := 0
				// hasIndustrialReports := report4Entry.HasIndustrialReports
				for _, childEntry := range childrenEntries {
					currChildEntry := reportMap[childEntry.Val.Path]
					fileCnt += currChildEntry.FileCnt
					pdfCnt += currChildEntry.PdfCnt
					pdfReadCnt += currChildEntry.PdfReadCnt
					// if currChildEntry.HasIndustrialReports {
					// 	hasIndustrialReports = true
					// }
				}
				report4Entry.FileCnt = fileCnt
				report4Entry.PdfCnt = pdfCnt
				report4Entry.PdfReadCnt = pdfReadCnt
				// report4Entry.HasIndustrialReports = hasIndustrialReports
				stack = stack[:len(stack)-1]
				continue
			}

			topEntryName := topEntry.Name()
			report4Entry := reportMap[parentPath]
			if topEntry.Type().IsRegular() {
				// case 1: 	是文件
				report4Entry.FileCnt = 1
				// case 2: 	而且是 pdf 文件
				if path.Ext(topEntryName) == ".pdf" {
					report4Entry.PdfCnt = 1
					if strings.Contains(topEntryName, "已读") || strings.Contains(topEntryName, "已阅") {
						report4Entry.IsRead = true
						report4Entry.PdfReadCnt = 1
					}
				}
			} else if topEntry.Type().IsDir() {
				// case 3: 	是目录
				childrenEntries := stackTop.Children
				fileCnt := 0
				pdfCnt := 0
				pdfReadCnt := 0
				// hasIndustrialReports := report4Entry.HasIndustrialReports
				for _, childEntry := range childrenEntries {
					currChildEntry := reportMap[childEntry.Val.Path]
					fileCnt += currChildEntry.FileCnt
					pdfCnt += currChildEntry.PdfCnt
					pdfReadCnt += currChildEntry.PdfReadCnt
					// if currChildEntry.HasIndustrialReports {
					// 	hasIndustrialReports = true
					// }
				}
				report4Entry.FileCnt = fileCnt
				report4Entry.PdfCnt = pdfCnt
				report4Entry.PdfReadCnt = pdfReadCnt
				// report4Entry.HasIndustrialReports = hasIndustrialReports
			} else {
				fmt.Printf("either IsDir nor IsRegular, current entry type:  %s\n", topEntry.Type())
			}

			// 出栈
			stack = stack[:len(stack)-1]
			continue
		}

		stackTop.Visited++
		currChildren := stackTop.Children
		nextChild := currChildren[stackTop.Visited]
		nextEntry := nextChild.Val.Entry
		entryName := nextEntry.Name()
		entryExt := path.Ext(entryName)
		currPath = fmt.Sprintf("%s\\%s", parentPath, entryName)
		isFile := nextEntry.Type().IsRegular()

		reportMap[currPath] = &GeneralizedReport{
			Name:       entryName,
			IsDir:      nextEntry.Type().IsDir(),
			IsFile:     isFile,
			IsRead:     false,
			Ext:        entryExt,
			IsPdf:      isFile && entryExt == ".pdf",
			PdfCnt:     0,
			PdfReadCnt: 0,
			FileCnt:    0,
			SelfPath:   currPath,
			Level:      reportMap[parentPath].Level + 1,
			ParentPath: parentPath,
			// HasIndustrialReports: entryName == "行业研报",
		}

		stack = append(stack, nextChild)
	}

	resReportList := []*GeneralizedReport{}
	for _, report := range reportMap {
		resReportList = append(resReportList, report)
	}
	// 	step 3: 	数据写入 db.
	//	TODO:	step 3
	return resReportList, nil
}

func genDirTreeBySys(rootPath string) (*MultiWayDirNodeV1, error) {
	defaultRoot := "C:\\Users\\Jiayi Xu\\Desktop\\研报"
	if rootPath == "" {
		rootPath = defaultRoot
	}

	rootEntries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}
	rootChildren := []*MultiWayDirNodeV1{}
	for _, rEntry := range rootEntries {
		rootChildren = append(rootChildren, &MultiWayDirNodeV1{
			Children: []*MultiWayDirNodeV1{},
			Visited:  -1,
			Val: &DirVal{
				Entry: rEntry,
				Path:  fmt.Sprintf("%s\\%s", rootPath, rEntry.Name()),
			},
		})
	}
	rootFolder := &MultiWayDirNodeV1{
		Children: rootChildren,
		Visited:  -1,
		Val: &DirVal{
			Entry: nil,
			Path:  rootPath,
		},
	}

	// 通过广度优先搜索, 构建以 rootFolder 为根节点的完整目录树结构
	q := []*MultiWayDirNodeV1{rootFolder}
	for len(q) > 0 {
		curr := q[0]
		for _, currChild := range curr.Children {
			childVal := currChild.Val
			if childVal.Entry.IsDir() {
				childPath := childVal.Path
				lv2Entries, err := os.ReadDir(childPath)
				if err != nil {
					return nil, err
				}
				lv2Children := []*MultiWayDirNodeV1{}
				for _, lv2Entry := range lv2Entries {
					lv2Children = append(lv2Children, &MultiWayDirNodeV1{
						Children: []*MultiWayDirNodeV1{},
						Visited:  -1,
						Val: &DirVal{
							Path:  fmt.Sprintf("%s\\%s", childPath, lv2Entry.Name()),
							Entry: lv2Entry,
						},
					})
				}
				currChild.Children = lv2Children
				q = append(q, &MultiWayDirNodeV1{
					Children: lv2Children,
					Visited:  -1,
					Val: &DirVal{
						Path:  childPath,
						Entry: childVal.Entry,
					},
				})
			}
		}
		q = q[1:]
	}

	return rootFolder, nil
}

func (rSvc *ReportService) InitReadStatusOfAllReports(rootPath string, ctx *gin.Context) ([]*GeneralizedReport, error) {
	// 直接遍历 "C:\\Users\\Jiayi Xu\\Desktop\\研报", 并生成所有文件夹下 pdf 记数的结果。跳过生成目录结构树这步
	defaultRoot := "C:\\Users\\Jiayi Xu\\Desktop\\研报"
	if rootPath == "" {
		rootPath = defaultRoot
	}

	rootFolder := &MultiWayDirNode{
		Children:   []*MultiWayDirNode{},
		Visited:    -1,
		ParentPath: "",
		Val: &DirVal{
			Entry: nil,
			Path:  rootPath,
		},
	}

	stack := []*MultiWayDirNode{rootFolder}
	var stackTop *MultiWayDirNode
	currLvPath := rootPath
	reportMap := make(map[string]*GeneralizedReport, 0)
	reportMap[currLvPath] = &GeneralizedReport{
		Name:              currLvPath,
		IsDir:             true,
		IsFile:            false,
		IsRead:            false,
		Ext:               "",
		IsPdf:             false,
		PdfReadCnt:        0,
		FileCnt:           0,
		PdfCnt:            0,
		SelfPath:          currLvPath,
		ParentPath:        "",
		Level:             0,
		IsIndustryFolder:  false,
		HasUnreadIndustry: false,
	}

	for len(stack) > 0 {
		stackTop = stack[len(stack)-1]
		topEntry := stackTop.Val.Entry
		lv2Children := []*MultiWayDirNode{}
		currLvPath = stackTop.Val.Path
		// step 1:	如果该节点未访问过, 则读取其所有子文件夹, 并放入 children
		if stackTop.Visited == -1 {
			// os.ReadDir 要求所搜索的路径必须是目录, 否则会抛出错误
			if topEntry == nil || topEntry.IsDir() {
				lv2Entries, err := os.ReadDir(currLvPath)
				if err != nil {
					return nil, err
				}
				for _, lv2Entry := range lv2Entries {
					lv2Children = append(lv2Children, &MultiWayDirNode{
						Children:   []*MultiWayDirNode{},
						Visited:    -1,
						ParentPath: currLvPath,
						Val: &DirVal{
							Entry: lv2Entry,
							Path:  fmt.Sprintf("%s\\%s", currLvPath, lv2Entry.Name()),
						},
					})
				}
				stackTop.Children = lv2Children
			}
		}

		// step 2:	加总所有子节点状态
		if len(stackTop.Children) == 0 || stackTop.Visited == len(stackTop.Children)-1 {
			// 1. 	如果是 叶子节点, 则更新其 pdf 阅读状态
			// 或
			// 2. 	所有子节点均已处理, 则加总其所有子节点的 pdf 阅读状态及是否"行业研报"状态
			report4Entry := reportMap[currLvPath]
			if topEntry == nil || topEntry.Type().IsDir() {
				// case 1: 	topEntry == nil 即 topEntry 是根节点的场景
				// case 2:  topEntry.Type().IsDir() 即当前 entry 是目录的场景
				childrenEntries := stackTop.Children
				fileCnt := 0
				pdfCnt := 0
				pdfReadCnt := 0
				hasChildUnreadIndustry := false
				for _, childEntry := range childrenEntries {
					currChildReport := reportMap[childEntry.Val.Path]
					fileCnt += currChildReport.FileCnt
					pdfCnt += currChildReport.PdfCnt
					pdfReadCnt += currChildReport.PdfReadCnt
					if currChildReport.HasUnreadIndustry {
						hasChildUnreadIndustry = true
					}
				}
				report4Entry.FileCnt = fileCnt
				report4Entry.PdfCnt = pdfCnt
				report4Entry.PdfReadCnt = pdfReadCnt
				if hasChildUnreadIndustry || (report4Entry.IsIndustryFolder && pdfCnt > pdfReadCnt) {
					report4Entry.HasUnreadIndustry = true
				}
			} else if topEntry.Type().IsRegular() {
				// case 3: 	topEntry 是文件的场景
				report4Entry.FileCnt = 1
				// case 4: 	而且 topEntry 是 pdf 文件
				topEntryName := topEntry.Name()
				if path.Ext(topEntryName) == ".pdf" {
					report4Entry.PdfCnt = 1
					if strings.Contains(topEntryName, "已读") || strings.Contains(topEntryName, "已阅") {
						report4Entry.IsRead = true
						report4Entry.PdfReadCnt = 1
					}
				}
			} else {
				fmt.Printf("either IsDir nor IsRegular, current entry type:  %s\n", topEntry.Type())
			}

			// 出栈
			stack = stack[:len(stack)-1]
			continue
		}

		//	step 3:	获取下一个入栈的节点, 并构造 GeneralizedReport, 作为结果集的一部分
		stackTop.Visited++
		nextChild := stackTop.Children[stackTop.Visited]
		nextEntry := nextChild.Val.Entry
		nextEntryName := nextEntry.Name()
		nextEntryExt := path.Ext(nextEntryName)
		nextPath := fmt.Sprintf("%s\\%s", currLvPath, nextEntryName)
		isFile := nextEntry.Type().IsRegular()
		isIndustryFolder := nextEntryName == "行业研报"

		reportMap[nextPath] = &GeneralizedReport{
			Name:              nextEntryName,
			IsDir:             nextEntry.Type().IsDir(),
			IsFile:            isFile,
			IsRead:            false,
			Ext:               nextEntryExt,
			IsPdf:             isFile && nextEntryExt == ".pdf",
			PdfCnt:            0,
			PdfReadCnt:        0,
			FileCnt:           0,
			SelfPath:          nextPath,
			Level:             reportMap[currLvPath].Level + 1,
			ParentPath:        currLvPath,
			IsIndustryFolder:  isIndustryFolder,
			HasUnreadIndustry: false,
		}

		stack = append(stack, nextChild)
	}

	// map 转化为数组作为结果集返回
	resReportList := []*GeneralizedReport{}
	for _, r := range reportMap {
		resReportList = append(resReportList, r)
	}
	return resReportList, nil
}
