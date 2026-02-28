# 12306 车票重命名工具（GUI）

从指定文件夹中扫描 12306 邮件下载的电子发票（`*.pdf` / `*.zip`），提取乘车日期与出发/到达站，按 `yyyy-mm-dd-出发站-到达站.pdf` 命名后输出到指定目录。

## 功能

- 支持单个 PDF 与批量 ZIP（ZIP 内可再嵌套 ZIP）
- 从 PDF 内嵌的 XBRL 提取：`TravelDate`、`DepartureStation`、`DestinationStation`（可切换用 `DateOfIssue`）
- 输出到指定目录，不修改输入目录的原文件
- 重名自动追加后缀：`-2`、`-3`…

## 使用说明

详见：`使用说明.txt`

## 开发构建（Windows）

```powershell
./build.ps1
```

输出：`dist/12306-invoice-renamer.exe`
