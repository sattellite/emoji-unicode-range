package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Emoji struct {
	Group  string
	Symbol string
	Desc   string
	Code   string
	Codes  []int
}

type EmojiGroup struct {
	Name    string  `json:"name"`
	Emojies []Emoji `json:"emojies"`
}

type Parser struct {
	List []EmojiGroup
}

func (p *Parser) Init(path string) {
	lines, err := p.ReadFile(path)
	if err != nil {
		panic(err)
	}
	p.List = p.Parse(lines)
}

func (p *Parser) ReadFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lines []string
	var isSuitable bool
	for scanner.Scan() {
		if !isSuitable && strings.Contains(scanner.Text(), "# group:") {
			isSuitable = true
		}

		if isSuitable {
			lines = append(lines, scanner.Text())
		}
	}

	return lines, nil
}

func (p *Parser) Parse(lines []string) []EmojiGroup {
	var groups []EmojiGroup
	var group EmojiGroup
	var groupName string
	for _, line := range lines {
		// Get new group
		if strings.Contains(line, "# group:") {
			groupName = p.parseGroupName(line)

			if len(group.Emojies) > 1 {
				groups = append(groups, group)
			}
			group = EmojiGroup{Name: groupName}
			continue
		}
		// Emoji parsing
		if strings.Contains(line, "; fully-qualified") {
			emoji := p.parseEmoji(line, "; fully-qualified")
			emoji.Group = groupName
			group.Emojies = append(group.Emojies, emoji)
		}

		// Component parsing
		if strings.Contains(line, "; component") {
			emoji := p.parseEmoji(line, "; component")
			emoji.Group = groupName
			group.Emojies = append(group.Emojies, emoji)
		}
	}
	// Add last group
	if len(group.Emojies) > 1 {
		groups = append(groups, group)
	}
	return groups
}

func (p *Parser) parseGroupName(line string) string {
	parts := strings.Split(line, ": ")
	return parts[1]
}

func (p *Parser) parseEmoji(line, sep string) Emoji {
	parts := strings.Split(line, sep)
	parts[0] = strings.Trim(parts[0], " ")
	parts[1] = strings.Trim(parts[1], " ")
	parts = append(parts, strings.Join(strings.Split(parts[1], " ")[2:], " "))
	parts[1] = strings.Split(parts[1], " ")[1]
	var codes []int

	for _, str := range strings.Split(parts[0], " ") {
		i, err := strconv.ParseInt(str, 16, 0)
		if err == nil {
			codes = append(codes, int(i))
		}
	}

	emoji := Emoji{
		Symbol: parts[1],
		Desc:   parts[2],
		Code:   parts[0],
		Codes:  codes}

	return emoji
}

func (p *Parser) GenerateJSON() string {
	structure := map[string]string{
		"Component":         "",
		"Smileys & People":  "people",
		"Smileys & Emotion": "people",
		"People & Body":     "people",
		"Animals & Nature":  "nature",
		"Food & Drink":      "food",
		"Travel & Places":   "travel",
		"Activities":        "activity",
		"Objects":           "objects",
		"Symbols":           "symbols",
		"Flags":             "flags",
	}

	preJSON := map[string][]string{
		"people":   []string{},
		"nature":   []string{},
		"food":     []string{},
		"travel":   []string{},
		"activity": []string{},
		"objects":  []string{},
		"symbols":  []string{},
		"flags":    []string{},
	}

	// Ordered JSON package https://gitlab.com/c0b/go-ordered-json
	order := []string{"people", "nature", "food", "travel", "activity", "objects", "symbols", "flags"}

	for _, gr := range p.List {
		if structure[gr.Name] == "" {
			continue
		}
		emojis := []string{}
		for _, emoji := range gr.Emojies {
			if !strings.Contains(emoji.Desc, "skin tone") {
				emojis = append(emojis, emoji.Symbol)
			}
		}
		preJSON[structure[gr.Name]] = append(preJSON[structure[gr.Name]], emojis...)
	}

	// Generate ordered JSON
	lastIndex := len(order) - 1
	jsonString := `{"emojis":{`

	for i := range order {
		jsonString = jsonString + `"` + order[i] + `":`
		j, _ := json.Marshal(preJSON[order[i]])
		jsonString = jsonString + string(j)
		if i != lastIndex {
			jsonString = jsonString + ","
		}
	}
	jsonString = jsonString + "}}"

	return jsonString
}

func (p *Parser) GenerateRange() string {
	var rawArray []int
	var step, deep int
	var hexRanges []string

	for _, gr := range p.List {
		for _, emoji := range gr.Emojies {
			rawArray = append(rawArray, emoji.Codes...)
		}
	}

	sort.Ints(rawArray)

	// Remove duplicates
	founded := make(map[int]bool)
	j := 0
	for i, x := range rawArray {
		if !founded[x] {
			founded[x] = true
			rawArray[j] = rawArray[i]
			j++
		}
	}
	rawArray = rawArray[:j]

	// Remove keycodes smaller 58
	j = 0
	for i, x := range rawArray {
		if x > 57 {
			rawArray[j] = rawArray[i]
			j++
		}
	}
	rawArray = rawArray[:j]

	rangeArr := make([][]int, 0)

	// Instead of first loop step
	rangeArr = append(rangeArr, []int{rawArray[0]})

	for i := 1; i < len(rawArray); i++ {
		if rawArray[i] == rangeArr[step][deep]+1 {
			rangeArr[step] = append(rangeArr[step], rawArray[i])
			deep++
			continue
		}
		rangeArr = append(rangeArr, []int{rawArray[i]})
		deep = 0
		step++
	}

	// Generate unicode ranges
	for i := range rangeArr {
		if len(rangeArr[i]) == 1 {
			hexRanges = append(hexRanges, fmt.Sprintf("U+%x", rangeArr[i][0]))
		}
		if len(rangeArr[i]) > 1 {
			hexRanges = append(hexRanges, fmt.Sprintf("U+%x-%x", rangeArr[i][0], rangeArr[i][len(rangeArr[i])-1]))
		}
	}

	return "unicode-range: " + strings.Join(hexRanges, ",")
}

func (p *Parser) GenerateStats() {
	var ranges []int
	var sum int
	var uniq int

	for _, gr := range p.List {
		fmt.Printf("%s: ", gr.Name)
		i := 0
		for _, emoji := range gr.Emojies {
			i++
			sum++
			if !strings.Contains(emoji.Desc, "skin tone") {
				uniq++
			}
			ranges = append(ranges, emoji.Codes...)
		}
		fmt.Printf(" %d emojies\n", i)
	}

	fmt.Printf("%d uniq emojies\n", uniq)
	fmt.Printf("Summary %d emojies\n", sum)
}

func checkArgs(args []string, keys map[string]bool) error {
	desc := "Generating emojis list and it unicode ranges"
	title := "Arguments:"
	emoji := "--emoji\tGenerating list of emojis"
	rng := "--range\tGenerating unicode ranges for CSS"
	stat := "--stats\tGenerate statistics about emojis"
	help := fmt.Errorf("%s\n%s\n\t%s\n\t%s\n\t%s", desc, title, emoji, rng, stat)

	if len(args) == 0 {
		return help
	}

	for i := range args {
		if args[i] == "--emoji" {
			keys["emoji"] = true
		}
		if args[i] == "--range" {
			keys["range"] = true
		}
		if args[i] == "--stats" {
			keys["stats"] = true
		}
	}

	if !keys["emoji"] && !keys["range"] && !keys["stats"] {
		return help
	}

	return nil
}

func main() {
	em := &Parser{}
	// You can get it here https://unicode.org/Public/emoji/12.0/emoji-test.txt or download other versions of emoji
	em.Init("./emoji-test.txt")

	args := os.Args[1:]
	is := make(map[string]bool)
	err := checkArgs(args, is)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	if is["emoji"] {
		fmt.Println(em.GenerateJSON())
	}

	if is["range"] {
		fmt.Println(em.GenerateRange())
	}

	if is["stats"] {
		em.GenerateStats()
	}
}
