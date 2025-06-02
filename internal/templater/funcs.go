package templater

import (
	"encoding/json"
	"fmt"
	"maps"
	"math/rand"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-faker/faker/v4"
	"github.com/gobeam/stringy"
	"github.com/google/uuid"
	"github.com/rs/xid"
	"mvdan.cc/sh/v3/shell"
	"mvdan.cc/sh/v3/syntax"

	sprig "github.com/go-task/slim-sprig/v3"
	"github.com/go-task/template"
)

var templateFuncs template.FuncMap

func init() {
	taskFuncs := template.FuncMap{
		"OS":     func() string { return runtime.GOOS },
		"ARCH":   func() string { return runtime.GOARCH },
		"numCPU": func() int { return runtime.NumCPU() },
		"catLines": func(s string) string {
			s = strings.ReplaceAll(s, "\r\n", " ")
			return strings.ReplaceAll(s, "\n", " ")
		},
		"splitLines": func(s string) []string {
			s = strings.ReplaceAll(s, "\r\n", "\n")
			return strings.Split(s, "\n")
		},
		"fromSlash": func(path string) string {
			return filepath.FromSlash(path)
		},
		"toSlash": func(path string) string {
			return filepath.ToSlash(path)
		},
		"exeExt": func() string {
			if runtime.GOOS == "windows" {
				return ".exe"
			}
			return ""
		},
		"shellQuote": func(str string) (string, error) {
			return syntax.Quote(str, syntax.LangBash)
		},
		"splitArgs": func(s string) ([]string, error) {
			return shell.Fields(s, nil)
		},
		// IsSH is deprecated.
		"IsSH": func() bool { return true },
		"joinPath": func(elem ...string) string {
			return filepath.Join(elem...)
		},
		"relPath": func(basePath, targetPath string) (string, error) {
			return filepath.Rel(basePath, targetPath)
		},
		"merge": func(base map[string]any, v ...map[string]any) map[string]any {
			cap := len(v)
			for _, m := range v {
				cap += len(m)
			}
			result := make(map[string]any, cap)
			maps.Copy(result, base)
			for _, m := range v {
				maps.Copy(result, m)
			}
			return result
		},
		"spew": func(v any) string {
			return spew.Sdump(v)
		},
	}

	// aliases
	taskFuncs["q"] = taskFuncs["shellQuote"]

	// Deprecated aliases for renamed functions.
	taskFuncs["FromSlash"] = taskFuncs["fromSlash"]
	taskFuncs["ToSlash"] = taskFuncs["toSlash"]
	taskFuncs["ExeExt"] = taskFuncs["exeExt"]

	idFuncs := template.FuncMap{
		"uuid": func() string {
			return uuid.New().String()
		},
		"xid": func() string {
			return xid.New().String()
		},
	}

	stringyFuncs := template.FuncMap{
		"camelCase": func(v string, rule ...string) string {
			return stringy.New(v).CamelCase(rule...).Get()
		},
		"snakeCase": func(v string, rule ...string) string {
			return stringy.New(v).SnakeCase(rule...).Get()
		},
		"pascalCase": func(v string, rule ...string) string {
			return stringy.New(v).PascalCase(rule...).Get()
		},
		"kebabCase": func(v string, rule ...string) string {
			return stringy.New(v).KebabCase(rule...).Get()
		},
		"upperFirst": func(v string) string {
			return stringy.New(v).UcFirst()
		},
		"lowerFirst": func(v string) string {
			return stringy.New(v).LcFirst()
		},
		"pad": func(v string, length int, with, padType string) string {
			return stringy.New(v).Pad(length, with, padType)
		},
		"acronym": func(v string) string {
			return stringy.New(v).Acronym().Get()
		},
	}

	fakerFuncs := template.FuncMap{
		// Address
		"_latitude": func() float64 {
			return faker.Latitude()
		},
		"_longitude": func() float64 {
			return faker.Longitude()
		},
		"_getRealAddress": func() map[string]any {
			b, _ := json.Marshal(faker.GetRealAddress())
			var m map[string]any
			_ = json.Unmarshal(b, &m)
			return m
		},

		// Datetime
		"_unixTime": func() int64 {
			return faker.UnixTime()
		},
		"_date": func() string {
			return faker.Date()
		},
		"_timeString": func() string {
			return faker.TimeString()
		},
		"_monthName": func() string {
			return faker.MonthName()
		},
		"_yearString": func() string {
			return faker.YearString()
		},
		"_dayOfWeek": func() string {
			return faker.DayOfWeek()
		},
		"_dayOfMonth": func() string {
			return faker.DayOfMonth()
		},
		"_timestamp": func() string {
			return faker.Timestamp()
		},
		"_century": func() string {
			return faker.Century()
		},
		"_timezone": func() string {
			return faker.Timezone()
		},
		"_timeperiod": func() string {
			return faker.Timeperiod()
		},

		// Internet
		"_email": func() string {
			return faker.Email()
		},
		"_macAddress": func() string {
			return faker.MacAddress()
		},
		"_domainName": func() string {
			return faker.DomainName()
		},
		"_url": func() string {
			return faker.URL()
		},
		"_username": func() string {
			return faker.Username()
		},
		"_ipv4": func() string {
			return faker.IPv4()
		},
		"_ipv6": func() string {
			return faker.IPv6()
		},
		"_password": func() string {
			return faker.Password()
		},

		// Words and Sentences
		"_word": func() string {
			return faker.Word()
		},
		"_sentence": func() string {
			return faker.Sentence()
		},
		"_paragraph": func() string {
			return faker.Paragraph()
		},

		// Payment
		"_creditCardType": func() string {
			return faker.CCType()
		},
		"_creditCardNumber": func() string {
			return faker.CCNumber()
		},
		"_currency": func() string {
			return faker.Currency()
		},
		"_amountWithCurrency": func() string {
			return faker.AmountWithCurrency()
		},

		// Person
		"_titleMale": func() string {
			return faker.TitleMale()
		},
		"_titleFemale": func() string {
			return faker.TitleFemale()
		},
		"_firstName": func() string {
			return faker.FirstName()
		},
		"_firstNameMale": func() string {
			return faker.FirstNameMale()
		},
		"_firstNameFemale": func() string {
			return faker.FirstNameFemale()
		},
		"_lastName": func() string {
			return faker.LastName()
		},
		"_name": func() string {
			if rand.Intn(100) > 50 {
				return fmt.Sprintf("%s %s", faker.FirstNameFemale(), faker.LastName())
			}
			return fmt.Sprintf("%s %s", faker.FirstNameMale(), faker.LastName())
		},

		// Phone
		"_phonenumber": func() string {
			return faker.Phonenumber()
		},
		"_tollFreePhoneNumber": func() string {
			return faker.TollFreePhoneNumber()
		},
		"_e164PhoneNumber": func() string {
			return faker.E164PhoneNumber()
		},
	}

	fakerFuncs["_CCType"] = fakerFuncs["_creditCardType"]
	fakerFuncs["_CCNumber"] = fakerFuncs["_creditCardNumber"]

	templateFuncs = template.FuncMap(sprig.TxtFuncMap())
	maps.Copy(templateFuncs, taskFuncs)
	maps.Copy(templateFuncs, idFuncs)
	maps.Copy(templateFuncs, stringyFuncs)
	maps.Copy(templateFuncs, fakerFuncs)
}
