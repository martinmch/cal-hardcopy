package main

import (
    "text/template"
    "os"
    "io"
    "fmt"
    "time"
    "strings"
    "strconv"

)

func check(err error) {
    if err != nil {
        fmt.Println("Error : %s", err.Error())
        os.Exit(1)
    }
}

type SpreadHeader struct {
    LeftYear string
    LeftMonth string
    LeftWeek string
    RightYear string
    RightMonth string
    RightWeek string
    FirstWeek HeaderWeek
    SecondWeek HeaderWeek
    ThirdWeek HeaderWeek
    RemainingDays HeaderWeek
}

type HeaderWeek struct {
    Days []HeaderDay
    WeekNumber int
    Highlight bool
}

type HeaderDay struct {
    DayOfWeek string
    DayOfWeekNumber string
}

type Week struct {
    Header SpreadHeader
    Monday  Weekday
    Tuesday  Weekday
    Wednesday  Weekday
    Thursday  Weekday
    Friday  Weekday
    Saturday  Weekday
    Sunday  Weekday
}

type Weekday struct {
    Num     string
    Name    string
    Event1  string
    Event2  string
    Event3  string
    Event4  string
}

func DateToPaddedLatexDate(d time.Time) string {
    day := strconv.Itoa(d.Day())
    if(len(day) == 1) {
        return `\phantom{0}` + day
    } else {
        return day
    }

}

func TranslateWeekday(w time.Weekday) string {
    switch w {
        case time.Sunday:
            return "Søndag"
        case time.Monday:
            return "Mandag"
        case time.Tuesday:
            return "Tirsdag"
        case time.Wednesday:
            return "Onsdag"
        case time.Thursday:
            return "Torsdag"
        case time.Friday:
            return "Fredag"
        case time.Saturday:
            return "Lørdag"
        default:
            return w.String()
    }
}

func TranslateMonth(m time.Month) string {
    switch m {
        case 1:
            return "Januar"
        case 2:
            return "Februar"
        case 3:
            return "Marts"
        case 4:
            return "April"
        case 5:
            return "Maj"
        case 6:
            return "Juni"
        case 7:
            return "Juli"
        case 8:
            return "August"
        case 9:
            return "September"
        case 10:
            return "Oktober"
        case 11:
            return "November"
        case 12:
            return "December"
        default:
            return m.String()
    }
}

func GetMondayOfLastWeek(year int) int {
    firstDayOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.Now().Location())
    // 1 (mon) => -7
    // 2 (tue) => -8
    // 3 (wed  => -9
    // 4 (thu) => -10
    // 5 (fri) => -11
    // 6 (sat) => -12
    // 7 (sun) => -13
    return (-int(firstDayOfYear.Weekday()) + 6)
}

func PrintIndent(msg string, indent int) {
    fmt.Println(strings.Repeat(" ", 4*indent), msg)
}

func WeekFileName(w Week) string {
    return "week" + w.Header.LeftWeek + "-" + w.Header.LeftYear + ".tex"
}

func SaveWeek(w Week) {
    tpl := "spread3.tpl.tex"
    file, err := os.Create("build/" + WeekFileName(w))
    check(err)
    spreadTemplate, err := template.New(tpl).Delims("[[", "]]").ParseFiles(tpl)
    check(err)
    defer file.Close()
    spreadTemplate.Execute(file, w)
}

func main() {
    year := 2022
    firstDayOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.Now().Location())
    fmt.Println("First day of 2022: " + firstDayOfYear.String() + " (" + firstDayOfYear.Weekday().String() + ")")

    daysToLastMonday := (int(firstDayOfYear.Weekday())+6) * -1
    firstMonday := firstDayOfYear.AddDate(0, 0, daysToLastMonday)
    fmt.Println("Last monday of 2021: " + firstMonday.String() + " (" + firstMonday.Weekday().String() + ")")

    firstDayOfNextYear := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Now().Location())
    fmt.Println("First day of 2023: " + firstDayOfNextYear.String() + " (" + firstDayOfNextYear.Weekday().String() + ")")

    // First day of year => Days till sunday
    // 1 (mon) => 6
    // 2 (tue) => 5
    // 3 wed   => 4
    // 4 thu  =>  3
    // 5 fri  =>  2 
    // 6 sat  =>  1 
    // 7 sun  =>  0
    daysToFirstSunday := (7 - int(firstDayOfYear.Weekday()) + 6)
    firstSundayOfNextYear := firstDayOfNextYear.AddDate(0, 0, daysToFirstSunday)
    // When the loop hits this monday, it should stop.
    endingMonday := firstSundayOfNextYear.AddDate(0, 0, 1)

    fmt.Println(strings.Repeat("=", 80))

    allWeeks := []Week{}

    weekMonday := firstMonday
    for weekMonday.Before(endingMonday) {
        year, weekNumber  := weekMonday.ISOWeek()
        week := Week{}

        fmt.Println("Week: ", weekNumber)
        fmt.Println("Year: ", year)
        weekSunday := weekMonday.AddDate(0, 0, 6)
        PrintIndent("Monday: " + weekMonday.String(), 1)
        PrintIndent("Sunday: " + weekSunday.String(), 1)
        // Get month string or combined month string
        isOverlapWeek := weekMonday.Month() != weekSunday.Month()
        isOverlapYear := weekMonday.Year() != weekSunday.Year()
        _, mondayWeek := weekMonday.ISOWeek()
        _, sundayWeek := weekSunday.ISOWeek()

        leftYear := strconv.Itoa(weekMonday.Year())
        leftMonth := TranslateMonth(weekMonday.Month())
        leftWeek := strconv.Itoa(mondayWeek)
        rightYear := strconv.Itoa(weekSunday.Year())
        rightMonth := TranslateMonth(weekSunday.Month())
        rightWeek := strconv.Itoa(sundayWeek)
        highlightFirstWeek := false
        highlightSecondWeek := false
        highlightThirdWeek := false
        var headerFirstWeek []HeaderDay
        var headerSecondWeek []HeaderDay
        var headerThirdWeek []HeaderDay
        var headerRemainingDays []HeaderDay
        // 7 days behind + current monday + 23 days forward = 31 days.
        headerStartDay := weekMonday.AddDate(0, 0, -7)
        // We need to add 24 days (instead of 23), to ensure Time.Before()
        // [strictly less than] holds.
        headerEndDay := weekMonday.AddDate(0, 0, 24)

        if(isOverlapWeek) {
            PrintIndent("Week overlaps month", 1)
            // Is overlap on left or right page?
            // Before thursday, but after sunday. Left page contains
            // mon, tue, wed
            overlappingDay := time.Date(weekSunday.Year(), weekSunday.Month(), 1, 0, 0, 0, 0, time.Now().Location()).Weekday()
            isLeftPage := overlappingDay < 4 && overlappingDay > 0
            if(isLeftPage) {
                leftMonth = leftMonth[:3] + " - " + rightMonth[:3]
            } else {
                rightMonth = leftMonth[:3] + " - " + rightMonth[:3]
            }
        }

        if(isOverlapYear) {
            overlappingDay := time.Date(weekSunday.Year(), weekSunday.Month(), 1, 0, 0, 0, 0, time.Now().Location()).Weekday()
            isLeftPage := overlappingDay < 4 && overlappingDay > 0
            if(isLeftPage) {
                // 20XX - 20XY
                leftYear = leftYear + " - " + strconv.Itoa(weekSunday.Year())
            } else {
                // 20XY - 20XX
                rightYear = leftYear + " - " + strconv.Itoa(weekSunday.Year())
            }
        }

        headerDate := headerStartDay
        count := 0
        for headerDate.Before(headerEndDay) {
            count = count + 1
            day := HeaderDay {
                DayOfWeek: TranslateWeekday(headerDate.Weekday())[:1],
                DayOfWeekNumber: strconv.Itoa(headerDate.Day()),
            }
            if(headerDate.Weekday() == time.Sunday) {
                day.DayOfWeek = "\\textbf{" + day.DayOfWeek + "}"
                day.DayOfWeekNumber = "\\textbf{" + day.DayOfWeekNumber + "}"
            }

            if (count < 8) {
                if(headerDate == weekMonday) {
                    highlightFirstWeek = true
                }
                headerFirstWeek = append(headerFirstWeek, day)
            } else if(count< 15) {
                if(headerDate == weekMonday){
                    highlightSecondWeek = true
                }
                headerSecondWeek = append(headerSecondWeek, day)
            } else if(count< 22) {
                if(headerDate == weekMonday){
                    highlightThirdWeek = true
                }
                headerThirdWeek = append(headerThirdWeek, day)
            } else if(count< 32) {
                headerRemainingDays = append(headerRemainingDays, day)
            }
            headerDate = headerDate.AddDate(0, 0, 1)
        }
        // Print the contents of the arrays.

        week.Header = SpreadHeader {
            LeftYear: leftYear,
            LeftMonth: leftMonth,
            LeftWeek: leftWeek,
            RightYear: rightYear,
            RightMonth: rightMonth,
            RightWeek: rightWeek,
            FirstWeek: HeaderWeek {
                Days: headerFirstWeek,
                Highlight: highlightFirstWeek,
            },
            SecondWeek: HeaderWeek {
                Days: headerSecondWeek,
                Highlight: highlightSecondWeek,
            },
            ThirdWeek: HeaderWeek {
                Days: headerThirdWeek,
                Highlight: highlightThirdWeek,
            },
            RemainingDays: HeaderWeek {
                Days: headerRemainingDays,
                Highlight: false,
            },
        }

        weekDay := weekMonday
        for weekDay.Before(weekSunday.AddDate(0, 0, 1)) {
            daydigit := weekDay.Weekday()
            d := Weekday {
                Num: DateToPaddedLatexDate(weekDay),
                Name: TranslateWeekday(daydigit),
                Event1: "",
                Event2: "",
                Event3: "",
                Event4: "",
            }
            switch daydigit {
                case 1:
                    week.Monday = d
                case 2:
                    week.Tuesday = d
                case 3:
                    week.Wednesday = d
                case 4:
                    week.Thursday = d
                case 5:
                    week.Friday = d
                case 6:
                    week.Saturday = d
                case 0:
                    week.Sunday = d
                default:
                    fmt.Printf("Didn't hit case for daydigit")
            }
            weekDay = weekDay.AddDate(0, 0, 1)
        }


        fmt.Printf("Week: %+v\n", week)
        allWeeks = append(allWeeks, week)
        weekMonday = weekMonday.AddDate(0, 0, 7)
    }

    srcFile, err := os.Open("main.tpl.tex")
    check(err)
    defer srcFile.Close()

    destFile, err := os.Create("build/main.tex") // creates if file doesn't exist
    check(err)
    defer destFile.Close()

    _, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
    check(err)

    err = destFile.Sync()
    check(err)

    // Writes main.tex
    f, err := os.OpenFile("build/main.tex",
    os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    check(err)
    defer f.Close()
    for _, w := range allWeeks {
        SaveWeek(w)

        f.WriteString("\\newpage\n\\input{" + WeekFileName(w) + "}\n")
    }
    f.WriteString("\\end{document}")
}

