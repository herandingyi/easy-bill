package currency

import (
	"errors"
	"math"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var CurrencyReg = []*regexp.Regexp{nil,
	regexp.MustCompile(`^[0-9]+\.?[0-9]{0,2}$`),
	regexp.MustCompile(`^[0-9]+$`),
	regexp.MustCompile(`^[0-9]+\.?[0-9]{0,2}$`),
}

var CurrencyDecimalPlace = []int{0,
	2,
	0,
	2,
}

var CurrencyName = []string{"",
	"USD",
	"KHR",
	"CNY",
}

var CurrencyTokens = [][]string{{},
	{"u", "U", "usd", "USD", "刀", "$", "美元", "美金"},
	{"k", "K", "khr", "KHR", "៛", "瑞尔"},
	{"c", "C", "cny", "CNY", "¥", "人民币", "元"},
}

var MinCurrencyTokenMustSpecify = []int64{0,
	100, // 小于 1.00 USD 时必须指定货币类型
	100, // 小于 100  KHR 时必须指定货币类型
	100, // 小于 1.00 CNY 时必须指定货币类型
}

var DefaultCurrencyType = 2

func CheckMin(useDefaultCurrencyType bool, money *big.Rat, currencyType int) (err error) {
	if useDefaultCurrencyType {
		minCurrencyTokenMustSpecify := MinCurrencyTokenMustSpecify[currencyType]
		f, _ := money.Float64()
		if int64(f) < minCurrencyTokenMustSpecify {
			err = errors.New(CurrencyName[currencyType] +
				" 总金额小于" +
				MarshalCurrencyNumber(minCurrencyTokenMustSpecify, currencyType) +
				"，必须明确指明货币类型")
			return
		}
	}
	return
}
func Parse(name string) (currencyType int, remain string, useDefaultCurrencyType bool) {
	name = " " + strings.TrimSpace(strings.ToLower(name))
	var tokens []string
	var token string
	useDefaultCurrencyType = true
	for currencyType, tokens = range CurrencyTokens {
		if currencyType == 0 {
			continue
		}
		for _, v := range tokens {
			v = " " + v
			if strings.HasSuffix(name, v) {

				hit := false
				{
					index := len(name) - len(v) - 1
					for index >= 0 {
						if name[index] == ' ' {
							index--
						} else {
							break
						}
					}
					if index >= 0 {
						last := name[index]
						// rzt,pa, k 其中 k需要算作人名; rzt,pa k 其中 k需要算作货币; rzt,pa10 k 其中 k需要算作货币
						//last 是字母或是数字
						if (last >= 97 && last <= 122) || (last >= 48 && last <= 57) {
							hit = true
						}
					} else {
						hit = true
					}
				}
				if hit {
					useDefaultCurrencyType = false
					token = v
					break
				}
			}
		}
		if token != "" {
			break
		}
	}
	if token == "" {
		currencyType = DefaultCurrencyType
		remain = name
	} else {
		remain = strings.TrimSuffix(name, token)
	}
	return
}

func GetSortCurrency() (tokens []int) {
	tokenArray := make([]string, 0, len(CurrencyTokens)-1)
	tokenMap := map[string]int{}
	for i, token := range CurrencyTokens {
		if i == 0 {
			continue
		}
		tokenArray = append(tokenArray, token[0])
		tokenMap[token[0]] = i
	}
	sort.Strings(tokenArray)
	for _, token := range tokenArray {
		tokens = append(tokens, tokenMap[token])
	}
	return
}

func MarshalCurrencyNumber(i int64, currencyType int) (num string) {
	decimalPlace := CurrencyDecimalPlace[currencyType]
	if decimalPlace == 0 {
		return strconv.FormatInt(i, 10)
	} else {
		denom := math.Pow(10, float64(decimalPlace))
		a := i / int64(denom)
		b := int64(math.Abs(float64(i))) % int64(denom)
		return strconv.FormatInt(a, 10) + "." + strconv.FormatInt(b, 10)
	}
}

func UnmarshalCurrencyNumber(num string, currencyType int) (a int64, err error) {
	num, err = TrimNumber(num)
	if err != nil {
		return 0, errors.New("金额非法")
	}
	_, err = strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, errors.New("金额必须是有效数字")
	}
	var reg = CurrencyReg[currencyType]
	if !reg.Match([]byte(num)) {
		return 0, errors.New("金额必须是有效数字 需满足的正则为：" + reg.String())
	}
	var decimalPlace = CurrencyDecimalPlace[currencyType]

	nums := strings.Split(num, ".")
	if len(nums) == 1 {
		nums = append(nums, "")
	}
	if len(nums) != 2 {
		return 0, errors.New("无法解析数据")
	}
	for i := len(nums[1]); i < decimalPlace; i++ {
		nums[1] += "0"
	}
	return strconv.ParseInt(nums[0]+nums[1], 10, 64)

}
func TrimNumber(num string) (ret string, err error) {
	if strings.Contains(num, ".") {
		return strings.TrimRight(strings.TrimSpace(num), "0"), nil
	} else {
		return strings.TrimLeft(strings.TrimSpace(num), "0"), nil
	}
}
