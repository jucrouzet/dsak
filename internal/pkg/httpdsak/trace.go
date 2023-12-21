package httpdsak

import (
	"github.com/fatih/color"
)

func (c *Client) doTrace(text string, attrs ...color.Attribute) {
	if !c.trace {
		return
	}
	printer := color.New(attrs...).FprintFunc()
	printer(c.log, text)
}

func (c *Client) doTracef(format string, args []any, attrs ...color.Attribute) {
	if !c.trace {
		return
	}
	printer := color.New(attrs...).FprintfFunc()
	printer(c.log, format, args...)
}

func (c *Client) traceInfo(text string) {
	c.doTrace(text, color.FgBlue)
}

func (c *Client) traceInfoln(text string) {
	c.doTrace(text+"\n", color.FgBlue)
}

func (c *Client) traceInfof(format string, a ...any) {
	c.doTracef(format, a, color.FgBlue)
}

func (c *Client) traceValue(text string) {
	c.doTrace(text, color.FgCyan)
}

func (c *Client) traceValueln(text string) {
	c.doTrace(text+"\n", color.FgCyan)
}

func (c *Client) traceValuef(format string, a ...any) {
	c.doTracef(format, a, color.FgCyan)
}

func (c *Client) traceInfoValue(info, value string) {
	c.traceInfof("%s: ", info)
	c.traceValue(value)
}

func (c *Client) traceInfoValuef(info, format string, a ...any) {
	c.traceInfof("%s: ", info)
	c.traceValuef(format, a...)
}

func (c *Client) traceInfoValueln(info, value string) {
	c.traceInfoValue(info, value)
	c.traceInfoln("")
}

func (c *Client) traceError(text string) {
	c.doTrace(text, color.FgRed)
}
func (c *Client) traceErrorln(text string) {
	c.doTrace(text+"\n", color.FgRed)
}

func (c *Client) traceErrorf(format string, a ...any) {
	c.doTracef(format, a, color.FgRed)
}

func (c *Client) traceErrorE(err error) {
	if err == nil {
		c.traceError("no error")
		return
	}
	c.traceError(err.Error())
}

func (c *Client) traceErrorEln(err error) {
	if err == nil {
		c.traceErrorln("no error")
		return
	}
	c.traceErrorln(err.Error())
}
