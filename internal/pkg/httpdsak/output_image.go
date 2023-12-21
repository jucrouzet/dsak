package httpdsak

import (
	"context"
	"errors"
	"fmt"
	"image"
	"net/http"
	"strconv"

	// Image formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/BourgeoisBear/rasterm"
)

func (c *Client) outputImage(ctx context.Context, res *http.Response) error {
	img, format, err := image.Decode(res.Body)
	if err != nil {
		return fmt.Errorf("failed to decode image : %w", err)
	}
	c.traceInfo("Content-Type : ")
	c.traceValueln(res.Header.Get("content-type"))

	c.traceInfo("Format : ")
	c.traceValueln(format)

	c.traceInfo("Dimensions : ")
	c.traceValuef("%dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())
	if res.Header.Get("content-length") != "" {
		s, err := strconv.ParseInt(res.Header.Get("content-length"), 10, 64)
		if err == nil {
			c.traceInfo("Size : ")
			c.traceValuef("%d (%s)\n", s, bytesSI(s))
		}
	}
	return c.outputImageTerm(ctx, img)
}

func (c *Client) outputImageTerm(_ context.Context, img image.Image) error {
	var err error
	settings := rasterm.Settings{
		EscapeTmux: false,
	}
	if rasterm.IsTermKitty() { //nolint:nestif
		err = settings.KittyWriteImage(c.out, img)
	} else if rasterm.IsTermItermWez() {
		err = settings.ItermWriteImage(c.out, img)
	} else {
		capable, capErr := rasterm.IsSixelCapable()
		if capErr != nil {
			return capErr
		}
		if !capable {
			return fmt.Errorf("terminal does not support sixel")
		}
		if iPaletted, ok := img.(*image.Paletted); ok {
			err = settings.SixelWriteImage(c.out, iPaletted)
		} else {
			err = errors.New("cannot use sixel output on non paletted images")
		}
	}
	return err
}

func bytesSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
