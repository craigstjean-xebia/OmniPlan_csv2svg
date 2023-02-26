package main

import (
	"embed"
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	BORDER_S     = 1
	W_PADDING_XS = 3
	W_PADDING_S  = 12
	W_PADDING    = 16
	W_MILESTONE  = 16
	H_PADDING_XS = 4
	H_PADDING_S  = 8
	H_PADDING    = 16
	H_DATESCALE  = 48
	H_DATETICK   = 12
	H_TODAY      = 5
	H_BAR        = 18
	H_TEXT       = 16
	RX           = 4
	RY           = 4

	DPI               = 72
	FONTFAMILY        = "Roboto"
	FONTSIZE_TIMELINE = 14
	FONTSIZE_BAR      = 12
	FONTSIZE_DATE     = 11
	TTF_BOLD          = "fonts/Roboto-Bold.ttf"
	TTF_REGULAR       = "fonts/Roboto-Regular.ttf"

	STYLE_BASE         = "background-color:white"
	STYLE_TODAYBAR     = "fill:#d4182d"
	STYLE_TODAYPARTBAR = "fill:orange"
	STYLE_TIMELINE     = "fill:#44536a"
	STYLE_BAR          = "fill:#589ad7"
	STYLE_MILESTONE    = "fill:#ecb22f"
	STYLE_DATETICK     = "stroke:white;stroke-width:2px"
	STYLE_BORDER       = "stroke:#f0f0f0;z-index:0"
)

var (
	STYLE_DATERANGETEXT = fmt.Sprintf("font:bold %dpx %s", FONTSIZE_TIMELINE, FONTFAMILY)
	STYLE_TIMELINETEXT  = fmt.Sprintf("font:bold %dpx %s;fill:white", FONTSIZE_TIMELINE, FONTFAMILY)
	STYLE_BARTEXT       = fmt.Sprintf("font:bold %dpx %s", FONTSIZE_BAR, FONTFAMILY)
	STYLE_DATETEXT      = fmt.Sprintf("font:normal %dpx %s", FONTSIZE_DATE, FONTFAMILY)
)

var (
	//go:embed fonts
	ttfs embed.FS
)

type Task struct {
	id            string
	title         string
	start         time.Time
	end           time.Time
	durationHours int
	effortHours   int
	completed     string
	assigned      string
	isMilestone   bool
}

type Tick struct {
	x1                   int
	x2                   int
	startDate            int64
	durationBusinessDays int
}

func getFontDrawer(ttf string, size int) *font.Drawer {
	dst := image.NewRGBA((image.Rect(0, 0, 1000, 1000)))

	b, err := ttfs.ReadFile(ttf)
	if err != nil {
		log.Fatalf("unable to read %s - %v\n", ttf, err)
	}

	f, err := truetype.Parse(b)
	if err != nil {
		log.Fatalf("unable to parse %s - %v\n", ttf, err)
	}

	return &font.Drawer{
		Dst: dst,
		Src: image.Black,
		Face: truetype.NewFace(f, &truetype.Options{
			Size: FONTSIZE_TIMELINE,
			DPI:  DPI,
		}),
		Dot: fixed.P(0, 0),
	}
}

func getTimelineFontDrawer() *font.Drawer {
	return getFontDrawer(TTF_BOLD, FONTSIZE_TIMELINE)
}

/*func getBarFontDrawer() *font.Drawer {
	return getFontDrawer(TTF_BOLD, FONTSIZE_BAR)
}*/

func getDateFontDrawer() *font.Drawer {
	return getFontDrawer(TTF_REGULAR, FONTSIZE_DATE)
}

func getWeekdaysBetween(start, end time.Time) int {
	offset := -int(start.Weekday())
	start = start.AddDate(0, 0, -int(start.Weekday()))

	offset += int(end.Weekday())
	if end.Weekday() == time.Sunday {
		offset++
	}

	end = end.AddDate(0, 0, -int(end.Weekday()))

	diff := end.Sub(start).Truncate(time.Hour * 24)
	weeks := float64((diff.Hours() / 24) / 7)
	return int(math.Round(weeks)*5) + offset
}

func adjustForWeekend(date time.Time) time.Time {
	if date.Weekday() == time.Saturday {
		date = date.AddDate(0, 0, 1)
	}
	if date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, 1)
	}

	return date
}

func addWeekdaysToDate(date time.Time, days float64) time.Time {
	currentDate := date
	daysInt := int(days)
	for i := 0; i < daysInt; i++ {
		currentDate = adjustForWeekend(currentDate.AddDate(0, 0, 1))
	}

	currentDate = adjustForWeekend(currentDate.Add(time.Minute * time.Duration(int(1440.0*(days-float64(daysInt))))))

	return currentDate
}

func readTaskData(csvFile string) []*Task {
	f, err := os.Open(csvFile)
	if err != nil {
		log.Fatalln("unable to read input file", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)

	_, err = reader.Read()
	if err != nil {
		log.Fatalln("unable to read input file", err)
	}

	const layout = "1/2/06, 3:04 PM"
	index := 2
	var tasks []*Task
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return tasks
			}

			log.Fatalln("unable to read input file", err)
		}

		t := &Task{}
		t.id = row[0]
		t.title = row[1]
		if row[2] != "" {
			if t.start, err = time.Parse(layout, row[2]); err != nil {
				log.Fatalf("unable to parse 'Start' of row %d (%s) - %v\n", index, row[2], err)
			}
		}
		if row[3] != "" {
			if t.end, err = time.Parse(layout, row[3]); err != nil {
				log.Fatalf("unable to parse 'End' of row %d (%s) - %v\n", index, row[3], err)
			}
		}
		if row[4] != "" {
			if t.durationHours, err = strconv.Atoi(row[4]); err != nil {
				log.Fatalf("unable to parse 'Duration Hours' of row %d (%s) - %v\n", index, row[4], err)
			}
		} else {
			t.isMilestone = true
		}
		if row[6] != "" {
			if t.effortHours, err = strconv.Atoi(row[6]); err != nil {
				log.Fatalf("unable to parse 'Effort Hours' of row %d (%s) - %v\n", index, row[6], err)
			}
		}
		t.completed = row[8]
		t.assigned = row[10]
		tasks = append(tasks, t)

		index++
	}
}

func getBarPositions(ticks []*Tick, start time.Time, end time.Time) (int, int) {
	x1 := 0
	x2 := 0

	durationDays := getWeekdaysBetween(start, end)
	if durationDays == 0 {
		durationDays++
	}

	for t := range ticks {
		tick := ticks[t]
		if tick.startDate == start.Truncate(time.Hour*24).Unix() {
			x1 = tick.x1
			x2 = x1 + ((tick.x2-tick.x1)/tick.durationBusinessDays)*durationDays
			return x1, x2
		} else if tick.startDate > start.Truncate(time.Hour*24).Unix() {
			daysAdj := getWeekdaysBetween(start, time.Unix(tick.startDate, 0)) + 1
			x1 = tick.x1 - (((tick.x2 - tick.x1) / tick.durationBusinessDays) * daysAdj)
			x2 = x1 + ((tick.x2-tick.x1)/tick.durationBusinessDays)*durationDays
			return x1, x2
		}
	}

	if x1 == 0 {
		log.Fatalf("Cannot determine position for %v - %v\n", start, end)
	}

	return x1, x2
}

func main() {
	var oFlag = flag.String("o", "", "output file")
	var widthFlag = flag.Int("w", 0, "force a specific width")
	var heightFlag = flag.Int("h", 0, "force a specific height")
	var levelFlag = flag.Int("level", 2, "maximum level to output")
	var zoomFlag = flag.String("zoom", "", "portion of ids to focus on, e.g. 1.4.1")
	var tFlag = flag.Int("t", 1, "number of days per tick mark")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatalln("input file required")
	}

	tasks := readTaskData(args[0])

	// remove all tasks indented beyond *levelFlag and tasks not in *zoomFlag
	zoom := *zoomFlag
	scrubbedTasks := tasks[:0]
	for t := range tasks {
		if len(strings.Split(tasks[t].id, ".")) <= *levelFlag &&
			(zoom == "" || tasks[t].id == zoom || strings.HasPrefix(tasks[t].id, zoom+".")) {
			scrubbedTasks = append(scrubbedTasks, tasks[t])
		}
	}
	tasks = scrubbedTasks

	// get min/max dates for range
	var minDate *time.Time = nil
	var maxDate *time.Time = nil
	for _, t := range tasks {
		if minDate == nil || t.start.Before(*minDate) {
			minDate = &t.start
		}
		if maxDate == nil || t.end.After(*maxDate) {
			maxDate = &t.end
		}
	}

	maxDatePlusOne := (*maxDate).AddDate(0, 0, 1)
	maxDate = &maxDatePlusOne

	dateRange := getWeekdaysBetween(*minDate, *maxDate)

	width := *widthFlag
	height := *heightFlag
	tickSkips := *tFlag

	writer := os.Stdout
	if *oFlag != "" {
		f, err := os.OpenFile(*oFlag, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalln("unable to open output file", err)
		}
		defer f.Close()

		writer = f
	}

	actualWidth := width
	if width == 0 {
		dateFontDrawer := getDateFontDrawer()
		actualWidth = 1024 + W_PADDING + dateFontDrawer.MeasureString("May 22 - May 22").Ceil()
		width = 1024
	}

	if height == 0 {
		height = H_PADDING + H_DATESCALE + H_PADDING + len(tasks)*(H_TEXT+H_PADDING_XS+H_BAR+H_PADDING_S) + H_PADDING
	}

	canvas := svg.New(writer)
	canvas.Start(actualWidth, height, `style="`+STYLE_BASE+`"`)
	canvas.Def()
	canvas.Style("text/css", "@import url('https://fonts.googleapis.com/css?family=Roboto:400,100,100italic,300,300italic,400italic,500,500italic,700,700italic,900,900italic');")
	canvas.DefEnd()

	y := H_PADDING

	timelineFontDrawer := getTimelineFontDrawer()

	// draw timeline header
	canvas.Roundrect(W_PADDING, y, width-W_PADDING*2, H_DATESCALE, RX, RY, STYLE_TIMELINE)

	currentDate := *minDate
	dayWidth := (width - W_PADDING*2) / dateRange

	var tickPositions []*Tick
	tickPositions = append(tickPositions, &Tick{
		x1:                   0,
		x2:                   (dayWidth * tickSkips) - 1,
		startDate:            (*minDate).Truncate(time.Hour * 24).Unix(),
		durationBusinessDays: tickSkips,
	})

	canvas.Text(W_PADDING+W_PADDING_XS, y+H_DATESCALE-(H_DATESCALE-FONTSIZE_TIMELINE)/2-2, (*minDate).Format("Jan 2"), STYLE_TIMELINETEXT)
	canvas.Line(W_PADDING, y+H_DATESCALE, W_PADDING, height-y, STYLE_BORDER)
	labelXEnd := int(timelineFontDrawer.MeasureString((*minDate).Format("Jan 2")).Ceil()) + W_PADDING_S

	canvas.Text(width-W_PADDING-W_PADDING_XS-int(timelineFontDrawer.MeasureString((*maxDate).Format("Jan 2")).Ceil()), y+H_DATESCALE-(H_DATESCALE-FONTSIZE_TIMELINE)/2-2, (*maxDate).Format("Jan 2"), STYLE_TIMELINETEXT)
	canvas.Line(width-W_PADDING, y+H_DATESCALE, width-W_PADDING, height-y, STYLE_BORDER)
	labelXStop := width - W_PADDING - W_PADDING_XS - int(timelineFontDrawer.MeasureString((*maxDate).Format("Jan 2")).Ceil())

	for i := dayWidth * tickSkips; i < width-W_PADDING*3; i += (dayWidth * tickSkips) {
		currentDate = addWeekdaysToDate(currentDate, float64(tickSkips))

		// add label if we wont' overlap
		if i > labelXEnd+W_PADDING {
			label := currentDate.Format("Jan 2")
			labelWidth := int(timelineFontDrawer.MeasureString(label).Ceil())

			if i+labelWidth < labelXStop {
				canvas.Text(W_PADDING+i-labelWidth/2, y+H_DATESCALE-(H_DATESCALE-FONTSIZE_TIMELINE)/2-2, label, STYLE_TIMELINETEXT)
			}

			labelXEnd = i + labelWidth
		}

		tickPositions = append(tickPositions, &Tick{
			x1:                   i,
			x2:                   i + (dayWidth * tickSkips) - 1,
			startDate:            currentDate.Truncate(time.Hour * 24).Unix(),
			durationBusinessDays: tickSkips,
		})

		canvas.Line(W_PADDING+i, y+H_DATESCALE-(H_DATESCALE-FONTSIZE_TIMELINE)/2+H_PADDING_XS,
			W_PADDING+i, y+H_DATESCALE-(H_DATESCALE-FONTSIZE_TIMELINE)/2-2+H_DATETICK+H_PADDING_XS,
			STYLE_DATETICK)
		canvas.Line(W_PADDING+i, y+H_DATESCALE-(H_DATESCALE-FONTSIZE_TIMELINE)/2+H_PADDING_XS,
			W_PADDING+i, height-y,
			STYLE_BORDER)
	}
	y += H_DATESCALE + H_PADDING

	// draw today indicator
	todayPosition := 0
	todayPortion := 0
	if time.Now().After(*maxDate) {
		todayPosition = width - W_PADDING*2
	} else if time.Now().After(*minDate) {
		_, todayPosition = getBarPositions(tickPositions, *minDate, time.Now().AddDate(0, 0, -1))

		now := time.Now()
		if now.Weekday() != time.Saturday && now.Weekday() != time.Sunday {
			todayPortion = int((float64(time.Now().Hour()*60+time.Now().Minute()) / float64(1440)) * float64(dayWidth))
		}
	}
	canvas.Roundrect(W_PADDING, y-H_PADDING-H_TODAY, todayPosition, H_TODAY, RX, RY, STYLE_TODAYBAR)
	canvas.Rect(W_PADDING, y-H_PADDING-H_TODAY, todayPosition, H_TODAY/2, STYLE_TODAYBAR)
	canvas.Rect(W_PADDING*2, y-H_PADDING-H_TODAY, todayPosition-W_PADDING, H_TODAY, STYLE_TODAYBAR)
	canvas.Rect(W_PADDING+todayPosition, y-H_PADDING-H_TODAY, todayPortion, H_TODAY, STYLE_TODAYPARTBAR)

	// draw tasks
	for ti := range tasks {
		t := tasks[ti]

		start, end := getBarPositions(tickPositions, t.start, t.end)

		if !t.isMilestone {
			canvas.Text(W_PADDING+start, y+H_TEXT, t.id+" - "+t.title, STYLE_BARTEXT)
		} else {
			canvas.Text(W_PADDING+start-W_MILESTONE/2, y+H_TEXT, t.id+" - "+t.title, STYLE_BARTEXT)
		}
		y += H_TEXT + H_PADDING_XS

		if !t.isMilestone {
			canvas.Roundrect(W_PADDING+start, y, end-start, H_BAR, RX, RY, STYLE_BAR)
		} else {
			end = start + W_MILESTONE
			canvas.Polygon(
				[]int{W_PADDING + start - W_MILESTONE/2, W_PADDING + start, W_PADDING + start + W_MILESTONE/2, W_PADDING + start},
				[]int{y + H_BAR/2, y, y + H_BAR/2, y + H_BAR},
				STYLE_MILESTONE)
		}

		canvas.Text(W_PADDING+end+W_PADDING, y+H_TEXT-3, t.start.Format("Jan 2")+" - "+t.end.Format("Jan 2"), STYLE_DATETEXT)

		y += H_BAR + H_PADDING_S
	}

	canvas.End()
}
