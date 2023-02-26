# OmniPlan CSV2SVG

Crudely written tool to take an Export from [OmniPlan](https://www.omnigroup.com/omniplan/) and turn it into a SVG
for presentations.

I did not like the Report capabilities built-in to OmniPlan, it does not
output in a pretty manner, does not allow you to hide or show different parts
of you plan (unless you just collapse something in the main window), and it
outputs in dark mode if your system is set to dark mode.

Inspired by [Office Timeline](https://www.officetimeline.com/). Note, if you
prefer Office Timeline, you can export from OmniPlan in a MPP format and import
that into Office Timeline, but I still found it difficult to get it to do
exactly what I wanted. I was also looking for something open-source.

## Building

```
go build
```

## How to use it

1. Open your Project in OmniPlan
2. Select File->Export
3. Set Format to CSV
4. Run omniplan_csv2svg against the CSV

## Usage

```
Usage of ./omniplan_csv2svg:
  -h int
        force a specific height
  -w int
        force a specific width
  -level int
        maximum level to output (default 2)
  -t int
        number of days per tick mark (default 1)
  -zoom string
        portion of ids to focus on, e.g. 1.4.1
  -o string
        output file
```

## TODO

- [X] Parse OmniPlan CSV Export
- [X] Generate SVG
- [X] Option to limit depth of ids
- [X] Option to zoom in to a specific id and its children
- [X] Today indicator
- [X] Add milestones
- [ ] Refactor
- [ ] Remove uneeded fonts
- [ ] omnijs for OmniPlan automation (if omnijs can export and can execute something)
- [ ] Add additional layouts, fonts, etc.
- [ ] Web interface for easy generation?
- [ ] Color code bars depending on percent complete, overdue, etc.
- [ ] Option to show assignee
- [ ] Add bar sizes and labels for hours, not just days

## Maybe Someday

- [ ] Option to show different baselines?
    - Does OmniPlan export these?
- [ ] Additional charts for visualizing costs, resource utilization, etc.
- [ ] Add additional parsing options for exporting from other tools
- [ ] Additional output options (PNG, JPG, PPT?)
