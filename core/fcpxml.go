package core

import (
	"encoding/xml"
	"regexp"
	"srt2fcpxml/core/FcpXML"
	"srt2fcpxml/core/FcpXML/Library"
	"srt2fcpxml/core/FcpXML/Library/Event"
	"srt2fcpxml/core/FcpXML/Library/Event/Project"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine/Gap"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine/Gap/Title"
	"srt2fcpxml/core/FcpXML/Resources"
	"strings"

	"github.com/asticode/go-astisub"
)

func Srt2FcpXmlExport(projectName string, frameDuration interface{}, subtitles *astisub.Subtitles, width, height int) ([]byte, error) {
	fcpxml := FcpXML.New()
	res := Resources.NewResources()
	res.SetEffect(Resources.NewEffect())
	format := Resources.NewFormat().
		SetWidth(width).
		SetHeight(height).
		SetFrameRate(frameDuration).Render()
	res.SetFormat(format)
	fcpxml.SetResources(res)
	gap := Gap.NewGap(subtitles.Duration().Seconds())

	// For subtitles with styles
	combinedRegex := regexp.MustCompile(`<[^>]*>|\{\\an\d\}`) // match <font> & {\anX}

	for index, item := range subtitles.Items {
		textStyleDef := Title.NewTextStyleDef(index + 1)
		text := Title.NewContent(index+1, func(lines []astisub.Line) string {
			var os []string
			for _, l := range lines {
				line := combinedRegex.ReplaceAllString(l.String(), "")
				os = append(os, line)
			}
			return strings.Join(os, "\n")
		}(item.Lines))
		// Use subtitle text itself as name
		title := Title.NewTitle(text.TextStyle.Text, item.StartAt.Seconds(), item.EndAt.Seconds()).SetTextStyleDef(textStyleDef).SetText(text)
		title.AddParam(Title.NewParams("Position", "9999/999166631/999166633/1/100/101", "0 -450"))
		title.AddParam(Title.NewParams("Alignment", "9999/999166631/999166633/2/354/999169573/401", "1 (Center)"))
		title.AddParam(Title.NewParams("Flatten", "9999/999166631/999166633/2/351", "1"))
		gap.AddTitle(title)
	}

	spine := Spine.NewSpine().SetGap(gap)
	seq := Sequence.NewSequence(subtitles.Duration().Seconds()).SetSpine(spine)
	project := Project.NewProject(projectName).SetSequence(seq)
	event := Event.NewEvent().SetProject(project)
	library := Library.NewLibrary(projectName).SetEvent(event)
	fcpxml.SetLibrary(library)

	return xml.MarshalIndent(fcpxml, "", "    ")
}
