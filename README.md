# 12306 车票重命名工具（GUI）

从指定文件夹中扫描 12306 邮件下载的电子发票附件（`*.pdf` / `*.zip`，ZIP 内可再嵌套 ZIP），提取乘车日期与出发/到达站，按 `yyyy-mm-dd-出发站-到达站.pdf` 命名后输出到指定目录。

说明：仅处理 `*.pdf` 与 `*.zip`，不处理 `*.ofd`。

## 功能

- 支持单个 PDF 与批量 ZIP（ZIP 内可再嵌套 ZIP）
- 从 PDF 内嵌的 XBRL 提取：`TravelDate`、`DepartureStation`、`DestinationStation`（可切换用 `DateOfIssue`）
- 输出到指定目录，不修改输入目录的原文件
- 重名自动追加后缀：`-2`、`-3`…
- 同名 PDF 去重：当扫描目录/ZIP（含嵌套 ZIP）发现“文件名相同”的 PDF 时，仅处理一个，其余会输出 `SKIP` 日志
- 记住上次选择的输入/输出目录（exe 同目录生成 `settings.json`）
- 首次打开默认输入目录为 `.\input`、输出目录为 `.\output`（相对 exe 所在目录）；若不存在会提示创建

## 使用说明

详见：`使用说明.txt`

## 开发构建（Windows）

```powershell
./build.ps1
```

输出：`dist/12306-invoice-renamer.exe`
