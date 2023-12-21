package httpdsak

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"math"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	// Image formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (c *Client) outputMedia(ctx context.Context, res *http.Response) error {
	if !hasFFMpeg() {
		return c.outputRaw(ctx, res)
	}
	// Read 5MB of data and keep a copy for ffmpeg.
	ffprobeBuffer := new(bytes.Buffer)
	ffmpegBuffer := new(bytes.Buffer)
	_, err := io.CopyN(io.MultiWriter(ffprobeBuffer, ffmpegBuffer), res.Body, 5*1024*1024)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read response body : %w", err)
	}
	var infos *ffprobeResponse
	if hasFFProbe() {
		infos, err = c.traceMedia(ctx, ffprobeBuffer)
		if err != nil {
			return errors.New("media is not supported")
		}
		c.showMediaFormat(infos.Format)
		c.showMediaPrograms(infos.Programs)
		c.showMediaStreams(infos.Streams)
	}
	return c.outputMediaFFmpeg(ctx, ffmpegBuffer, infos)
}

func hasFFProbe() bool {
	_, err := exec.LookPath("ffprobe")
	return err == nil
}

func hasFFMpeg() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

type ffprobeResponseProgram struct {
	Tags map[string]string `json:"tags"`
}
type ffprobeResponseStream struct {
	AvgFrameRate       string         `json:"avg_frame_rate"`
	BitRate            string         `json:"bit_rate,omitempty"`
	BitsPerRawSample   string         `json:"bits_per_raw_sample,omitempty"`
	BitsPerSample      int            `json:"bits_per_sample,omitempty"`
	ChannelLayout      string         `json:"channel_layout,omitempty"`
	Channels           int            `json:"channels,omitempty"`
	ChromaLocation     string         `json:"chroma_location,omitempty"`
	ClosedCaptions     int            `json:"closed_captions,omitempty"`
	CodecLongName      string         `json:"codec_long_name"`
	CodecName          string         `json:"codec_name"`
	CodecTag           string         `json:"codec_tag"`
	CodecTagString     string         `json:"codec_tag_string"`
	CodecType          string         `json:"codec_type"`
	CodedHeight        int            `json:"coded_height,omitempty"`
	CodedWidth         int            `json:"coded_width,omitempty"`
	DisplayAspectRatio string         `json:"display_aspect_ratio,omitempty"`
	Disposition        map[string]int `json:"disposition"`
	ExtradataSize      int            `json:"extradata_size,omitempty"`
	FieldOrder         string         `json:"field_order,omitempty"`
	FilmGrain          int            `json:"film_grain,omitempty"`
	HasBFrames         int            `json:"has_b_frames,omitempty"`
	Height             int            `json:"height,omitempty"`
	ID                 string         `json:"id"`
	Index              int            `json:"index"`
	InitialPadding     int            `json:"initial_padding,omitempty"`
	IsAvc              string         `json:"is_avc,omitempty"`
	Level              int            `json:"level,omitempty"`
	NalLengthSize      string         `json:"nal_length_size,omitempty"`
	NbReadFrames       string         `json:"nb_read_frames"`
	PixFmt             string         `json:"pix_fmt,omitempty"`
	Profile            string         `json:"profile"`
	RFrameRate         string         `json:"r_frame_rate"`
	Refs               int            `json:"refs,omitempty"`
	SampleAspectRatio  string         `json:"sample_aspect_ratio,omitempty"`
	SampleFmt          string         `json:"sample_fmt,omitempty"`
	SampleRate         string         `json:"sample_rate,omitempty"`
	StartPts           int64          `json:"start_pts"`
	StartTime          string         `json:"start_time"`
	TimeBase           string         `json:"time_base"`
	TsPacketsize       string         `json:"ts_packetsize"`
	Width              int            `json:"width,omitempty"`
}

type ffprobeResponseFormat struct {
	Filename       string `json:"filename"`
	FormatLongName string `json:"format_long_name"`
	FormatName     string `json:"format_name"`
	NbPrograms     int    `json:"nb_programs"`
	NbStreams      int    `json:"nb_streams"`
	ProbeScore     int    `json:"probe_score"`
	Size           string `json:"size"`
	StartTime      string `json:"start_time"`
}

type ffprobeResponse struct {
	Format   *ffprobeResponseFormat    `json:"format"`
	Programs []*ffprobeResponseProgram `json:"programs"`
	Streams  []*ffprobeResponseStream  `json:"streams"`
}

func (c *Client) traceMedia(ctx context.Context, data io.Reader) (*ffprobeResponse, error) {
	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		"-print_format",
		"json",
		"-count_frames",
		"-show_format",
		"-show_streams",
		"-show_programs",
		"-",
	)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdin = data
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		c.traceInfoln("ffprobe output :")
		c.traceValueln(stderr.String())
		return nil, errors.New("failed to run ffprobe on media start")
	}
	var response *ffprobeResponse
	if err := json.NewDecoder(stdout).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode ffprobe response: %w", err)
	}
	return response, nil
}

func (c *Client) showMediaFormat(f *ffprobeResponseFormat) {
	if f.FormatName != "" {
		format := f.FormatName
		if f.FormatLongName != "" {
			format += fmt.Sprintf(" (%s)", f.FormatLongName)
		}
		c.traceInfoValueln("Format", format)
	}
}

func (c *Client) showMediaPrograms(prgs []*ffprobeResponseProgram) {
	for i, prg := range prgs {
		if len(prg.Tags) > 0 {
			c.traceInfoValuef("Program", "#%d\n", i)
			c.traceInfoln("\tTags:")
			for k, v := range prg.Tags {
				c.traceInfo("\t\t")
				c.traceInfoValueln(k, v)
			}
			c.traceInfoln("")
		}
	}
}
func (c *Client) showMediaStreams(streams []*ffprobeResponseStream) {
	for _, s := range streams {
		c.traceInfo("Stream #")
		c.traceValuef("%d (%s)", s.Index, s.CodecType)
		c.traceInfoln(":")
		codec := getCodecString(s)
		if codec != "" {
			c.traceInfo("\tCodec: ")
			c.traceValueln(codec)
		}
		switch s.CodecType {
		case "audio":
			c.showMediaStreamAudio(s)
		case "video":
			c.showMediaStreamVideo(s)
		}
		disps := make([]string, 0, len(s.Disposition))
		for k, v := range s.Disposition {
			if v != 0 {
				disps = append(disps, cases.Title(language.English, cases.Compact).String(strings.Replace(k, "_", " ", -1)))
			}
		}
		if len(disps) > 0 {
			c.traceInfoln("\tDispositions:")
			for _, disp := range disps {
				c.traceInfo("\t\t-")
				c.traceValueln(disp)
			}
		}

		c.traceInfoln("")
	}
}

func (c *Client) showMediaStreamAudio(s *ffprobeResponseStream) {
	if s.SampleRate != "" {
		sr, err := strconv.ParseInt(s.SampleRate, 10, 64)
		if err == nil {
			c.traceInfoValuef("\tSample rate", "%.2f kHz\n", float64(sr)/1000)
		}
	}
	if s.Channels != 0 {
		c.traceInfoValuef("\tChannels", "%d\n", s.Channels)
	}
	if s.ChannelLayout != "" {
		c.traceInfoValueln("\tChannel layout", s.ChannelLayout)
	}
	if s.BitRate != "" {
		br, err := strconv.ParseInt(s.BitRate, 10, 64)
		if err == nil {
			c.traceInfoValuef("\tBit rate", "%.2f kbps\n", float64(br)/1024)
		}
	}
}

func (c *Client) showMediaStreamVideo(s *ffprobeResponseStream) {
	if s.Width != 0 && s.Height != 0 {
		c.traceInfoValueln("\tResolution", fmt.Sprintf("%dx%d", s.Width, s.Height))
	}
	if s.DisplayAspectRatio != "" {
		c.traceInfoValueln("\tAspect ratio", s.DisplayAspectRatio)
	}
	if s.PixFmt != "" {
		c.traceInfoValueln("\tPixel format", s.PixFmt)
	}
	if s.AvgFrameRate != "" {
		parts := strings.SplitN(s.AvgFrameRate, "/", 2)
		if len(parts) == 2 {
			a, errA := strconv.ParseInt(parts[0], 10, 64)
			b, errB := strconv.ParseInt(parts[1], 10, 64)
			if errA == nil && errB == nil {
				c.traceInfoValuef("\tFrame rate", "%.2f fps\n", float64(a)/float64(b))
				goto afterFR
			}
		}
		c.traceInfoValueln("\tAverage framerate", s.AvgFrameRate)
	}
afterFR:
	if s.BitRate != "" {
		br, err := strconv.ParseInt(s.BitRate, 10, 64)
		if err == nil {
			c.traceInfoValuef("\tBit rate", "%.2f mbps\n", float64(br)/(1024*1024))
		}
	}
}

func (c *Client) outputMediaFFmpeg(ctx context.Context, data io.Reader, infos *ffprobeResponse) error {
	stream := -1
	frame := 0
	if infos != nil {
		for _, s := range infos.Streams {
			if s.CodecType == "video" && s.NbReadFrames != "" {
				frames, err := strconv.ParseInt(s.NbReadFrames, 10, 64)
				if err == nil {
					frame = int(math.Round(float64(frames) / 2))
					stream = s.Index
				}
			}
		}
	}
	args := []string{"-i", "-"}
	if stream >= 0 {
		args = append(args, "-map", fmt.Sprintf("0:%d", stream))
	}
	args = append(
		args,
		"-vf",
		fmt.Sprintf(`select=eq(n\,%d)`, frame),
		"-q:v",
		"3",
		"-frames:v",
		"1",
		"-c:v",
		"png",
		"-f",
		"image2",
		"-",
	)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdin = data
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		c.traceInfoln("ffmpeg output :")
		c.traceValueln(stderr.String())
		return errors.New("failed to run ffmpeg on media start")
	}
	img, _, err := image.Decode(stdout)
	if err != nil {
		return fmt.Errorf("failed to decode image : %w", err)
	}
	return c.outputImageTerm(ctx, img)
}

func getCodecString(s *ffprobeResponseStream) string {
	if s.CodecName == "" {
		return ""
	}
	codec := s.CodecName

	if s.CodecLongName != "" {
		codec += fmt.Sprintf(" (%s)", s.CodecLongName)
	}

	if s.CodecTagString != "" {
		codec += fmt.Sprintf(" Tag: %s", s.CodecTagString)
	} else if s.CodecTag != "" {
		codec += fmt.Sprintf(" Tag: %s", s.CodecTag)
	}

	if s.Profile != "" {
		codec += fmt.Sprintf(" Profile: %s", s.Profile)
	}
	return codec
}
