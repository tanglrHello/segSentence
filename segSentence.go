package main

import (
	//"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"
)

func main() {

	rawBlogDataDir := "raw_data/"
	seggedBlogDataDir := "segged_data/"
	rawBlogFiles, _ := ioutil.ReadDir(rawBlogDataDir)
	blogFileNum := len(rawBlogFiles)

	for _, fileInfo := range rawBlogFiles {
		fileName := fileInfo.Name()

		//open in raw data file
		blogfile, err := os.Open(rawBlogDataDir + "/" + fileName)
		if err != nil {
			fmt.Println("err when open file %s:%q", fileName, err)
			continue
		}

		//segment sentences
		text, _ := ioutil.ReadAll(blogfile)
		blogfile.Close()
		textstr := string(text)
		textstr = strings.Replace(textstr, "\r", "", -1)
		paras := strings.Split(textstr, "\n")
		realtext := strings.Join(paras[1:], "\n")
		//fmt.Println(realtext)
		sentences := segmentSentences(realtext)

		//write segment result into new file
		newfile, err := os.Create(seggedBlogDataDir + "/" + fileName)
		defer newfile.Close()
		if err != nil {
			fmt.Println("err when create file %s:%q", fileName, err)
			return
		}
		for _, sentence := range sentences {
			newfile.WriteString(sentence)
			newfile.WriteString("\n")
		}
		newfile.Close()
	}
	fmt.Println("blog file number:%d", blogFileNum)

	//text := "李肃：“听说他搬。到加州去了，是这样吗？”IC·史密斯：“我不能对此发表评论。”李肃：“甚至还有人说他在南美洲被中国特工暗杀了。”IC·史密斯：“我可以戳穿那样的说法。那不是真的。”李肃：“直到今天仍然是这样？”IC·史密斯：“过去几天就说不好了。（笑）我也看过那样的报道，说他在危地马拉还是什么地方。”李肃：“没错，在海边。所以那不是真的？”sdadads"
	// text := "《阿迪盛大的韩？山撒撒打算打算的》"
	// for _, s := range segmentSentences(text) {
	// 	fmt.Println(s, "**")
	// }
}

func segmentSentences(text string) []string {
	paragraphs := strings.Split(text, "\n")
	var segres []string

	//sentenceSeps := "。？！：；｛｝（）［］【】“”‘’'……<>{}[]()?!;:\"\n\r"
	sentenceSeps := "。？！；｛｝【】'……{}?!;\n\r 　" //two empty character in the end

	outContainers := []string{"《》", "\"\"", "“”"} //outContainers and sentenceSeps should be exclusive
	seps := []rune(sentenceSeps)

	for _, para := range paragraphs {
		if strings.TrimSpace(para) == "" {
			continue
		}

		//extend spaces for runes belonging to sentenceSeps
		extendSepsStr := "。？！?!"
		extendSeps := []rune(extendSepsStr)
		extendedpara := make([]rune, utf8.RuneCountInString(para))

		runepara := []rune(para)
		for i, character := range runepara {
			sepFlag := false
			for _, esep := range extendSeps {
				if character == esep && i < len(extendSeps)-1 && runepara[i+1] != rune('”') {
					sepFlag = true
					extendedpara = append(extendedpara, character)
					extendedpara = append(extendedpara, rune(' '))
					break
				}
			}
			if sepFlag == false {
				extendedpara = append(extendedpara, character)
			}
		}

		para = string(extendedpara)

		para = strings.Replace(para, "”", "” ", -1)
		para = strings.Replace(para, "\"", "\" ", -1)

		var flag rune
		flag = '&'

		runeArr := []rune(strings.TrimSpace(para))
		runeArr = append(runeArr, rune(' '))

		for i := 0; i < len(runeArr); i++ {

			//preprocess
			endsep := []rune{'。', '？', '！', '?', '!', '…'}
			if i < len(runeArr)-1 {
				if runeArr[i] == '：' && runeArr[i+1] == '“' {
					continue
				}
				if runeArr[i] == ':' && runeArr[i+1] == '"' {
					continue
				}
			}

			//determine whether this rune is in sep List
			noSepFlag := false
			//normal segment by seps
			for _, sep := range seps {
				if sep == runeArr[i] {
					//test all pairs of outside container flags for this sep
					for _, outPair := range outContainers {
						borders := strings.Split(outPair, "")
						leftChar := []rune(borders[0])[0]
						rightChar := []rune(borders[1])[0]

						leftPos := findTowardsLeft(runeArr, i, leftChar)
						if leftPos == -1 {
							continue
						}

						rightPos := findTowardsRight(runeArr, i, rightChar)
						if rightPos == -1 {
							continue
						}

						rightCharInLeft := findTowardsLeft(runeArr, i, rightChar)
						leftCharInRight := findTowardsRight(runeArr, i, leftChar)

						if rightCharInLeft > leftPos {
							continue
						}
						if leftCharInRight < rightPos && leftCharInRight != -1 {
							continue
						}
						noSepFlag = true
						break
					}
					if noSepFlag == false {
						if strings.IndexRune(extendSepsStr, sep) != -1 {
							runeArr[i+1] = flag
							i++
						} else {
							runeArr[i] = flag
						}
					}
					break
				}
			}

			//post process
			if i > 0 {
				for _, es := range endsep {
					//fmt.Println(string(es))
					if runeArr[i] == '”' && runeArr[i-1] == es {
						runeArr[i+1] = flag

					}
					if runeArr[i] == '"' && runeArr[i-1] == es {
						runeArr[i] = flag
					}
				}
			}

		}

		tmp := string(runeArr)
		tmp = strings.Replace(tmp, "  ", string(flag), -1)

		sentences := strings.Split(tmp, string(flag))
		//postprocess
		for _, sentence := range sentences {
			trimmedSentence := strings.TrimSpace(sentence)
			//drop empty sentence
			if trimmedSentence == "" {
				continue
			}
			//drop short sentence
			if utf8.RuneCountInString(trimmedSentence) < 5 {
				continue
			}
			segres = append(segres, trimmedSentence)
		}
	}
	return segres
}

func findTowardsLeft(text []rune, pos int, c rune) int {
	for i := pos; i >= 0; i-- {
		if text[i] == c {
			return i
		}
	}
	return -1
}

func findTowardsRight(text []rune, pos int, c rune) int {
	for i := pos; i < len(text); i++ {
		if text[i] == c {
			return i
		}
	}
	return -1
}
