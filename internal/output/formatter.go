package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/tidwall/gjson"
)

// Writer io.Writer 类型别名
type Writer = io.Writer

// Format 输出格式类型
type Format string

const (
	FormatTable     Format = "table"
	FormatJSON      Format = "json"
	FormatNDJSON    Format = "ndjson"
	FormatMarkdown  Format = "markdown"
)

// Formatter 输出格式化器
type Formatter struct {
	format Format
	writer io.Writer
	quiet  bool
}

// NewFormatter 创建格式化器
func NewFormatter(format string, output string, quiet bool) *Formatter {
	f := FormatTable
	switch strings.ToLower(format) {
	case "json":
		f = FormatJSON
	case "ndjson":
		f = FormatNDJSON
	case "markdown", "md":
		f = FormatMarkdown
	}

	w := os.Stdout
	if output != "" {
		file, err := os.Create(output)
		if err == nil {
			w = file
		}
	}

	return &Formatter{
		format: f,
		writer: w,
		quiet:  quiet,
	}
}

// Output 输出数据
func (f *Formatter) Output(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.outputJSON(data)
	case FormatNDJSON:
		return f.outputNDJSON(data)
	case FormatMarkdown:
		return f.outputMarkdown(data)
	default:
		return f.outputTable(data)
	}
}

// OutputJSON 输出 JSON 格式
func (f *Formatter) outputJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// OutputNDJSON 输出 NDJSON 格式（每行一个 JSON）
func (f *Formatter) outputNDJSON(data interface{}) error {
	// 如果是数组，逐行输出
	if arr, ok := data.([]interface{}); ok {
		for _, item := range arr {
			bytes, err := json.Marshal(item)
			if err != nil {
				return err
			}
			fmt.Fprintln(f.writer, string(bytes))
		}
		return nil
	}
	// 单个对象直接输出
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Fprintln(f.writer, string(bytes))
	return nil
}

// OutputMarkdown 输出 Markdown 格式
func (f *Formatter) outputMarkdown(data interface{}) error {
	// 简单实现：转换为 JSON 格式的代码块
	fmt.Fprintln(f.writer, "```json")
	encoder := json.NewEncoder(f.writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(data)
	fmt.Fprintln(f.writer, "```")
	return err
}

// OutputTable 输出表格格式
func (f *Formatter) outputTable(data interface{}) error {
	// 如果是 gjson.Result，转换为表格
	if result, ok := data.(gjson.Result); ok {
		return f.outputGJSONTable(result)
	}

	// 其他类型暂时输出 JSON
	return f.outputJSON(data)
}

// outputGJSONTable 从 gjson 结果输出表格
func (f *Formatter) outputGJSONTable(result gjson.Result) error {
	t := table.NewWriter()
	t.SetOutputMirror(f.writer)

	// 处理数组数据
	if result.IsArray() {
		// 从第一个元素获取列名
		if len(result.Array()) > 0 && result.Array()[0].IsObject() {
			first := result.Array()[0]
			headers := table.Row{}
			for key := range first.Map() {
				headers = append(headers, key)
			}
			t.AppendHeader(headers)

			// 添加数据行
			for _, item := range result.Array() {
				row := table.Row{}
				for _, h := range headers {
					val := item.Get(h.(string))
					row = append(row, formatGJSONValue(val))
				}
				t.AppendRow(row)
			}
		}
	} else if result.IsObject() {
		// 单个对象：键值表格
		t.AppendHeader(table.Row{"字段", "值"})
		for key, val := range result.Map() {
			t.AppendRow(table.Row{key, formatGJSONValue(val)})
		}
	}

	t.Render()
	return nil
}

// formatGJSONValue 格式化 gjson 值
func formatGJSONValue(val gjson.Result) string {
	if val.IsObject() || val.IsArray() {
		bytes, _ := json.Marshal(val.Value())
		return string(bytes)
	}
	return val.String()
}

// Print 打印文本（如果不是 quiet 模式）
func (f *Formatter) Print(msg string) {
	if !f.quiet {
		fmt.Fprint(f.writer, msg)
	}
}

// Println 打印文本行（如果不是 quiet 模式）
func (f *Formatter) Println(msg string) {
	if !f.quiet {
		fmt.Fprintln(f.writer, msg)
	}
}

// Printf 格式打印（如果不是 quiet 模式）
func (f *Formatter) Printf(format string, args ...interface{}) {
	if !f.quiet {
		fmt.Fprintf(f.writer, format, args...)
	}
}

// Close 关闭输出（如果是文件）
func (f *Formatter) Close() error {
	if closer, ok := f.writer.(io.Closer); ok && f.writer != os.Stdout {
		return closer.Close()
	}
	return nil
}

// Writer 获取底层 writer
func (f *Formatter) Writer() io.Writer {
	return f.writer
}

// TableWriter 简化的表格写入器
type TableWriter struct {
	t     table.Writer
	w     io.Writer
	rows  []table.Row
}

// NewTableWriter 创建表格写入器
func NewTableWriter(w io.Writer) *TableWriter {
	t := table.NewWriter()
	return &TableWriter{
		t: t,
		w: w,
	}
}

// AppendHeader 添加表头
func (tw *TableWriter) AppendHeader(headers ...string) {
	row := table.Row{}
	for _, h := range headers {
		row = append(row, h)
	}
	tw.t.AppendHeader(row)
}

// AppendRow 添加数据行
func (tw *TableWriter) AppendRow(values ...string) {
	row := table.Row{}
	for _, v := range values {
		row = append(row, v)
	}
	tw.t.AppendRow(row)
}

// Render 渲染表格
func (tw *TableWriter) Render() {
	tw.t.SetOutputMirror(tw.w)
	tw.t.Render()
}