package stringutils

import (
	//"regexp"
	//"regexp/syntax"
	"strings"
	"unicode"
)

// const (
//	UNDERSCORE_PATTERN_1, _ = regexp.Compile("([A-Z]+)([A-Z][a-z])")
//	UNDERSCORE_PATTERN_2    = Pattern.compile("([a-z\\d])([A-Z])")
// )

var (
	plurals      = make([]*RuleAndReplacement, 0, 20)
	singulars    = make([]*RuleAndReplacement, 0, 20)
	uncountables = make([]string, 0, 20)
)

type RuleAndReplacement struct {
	rule        string
	replacement string
}

func init() {

	irregular("move", "moves")
	irregular("sex", "sexes")
	irregular("child", "children")
	irregular("man", "men")
	irregular("person", "people")

	add_plural("(quiz)$", "$1zes")
	add_plural("(ox)$", "$1en")
	add_plural("([m|l])ouse$", "$1ice")
	add_plural("(matr|vert|ind)ix|ex$", "$1ices")
	add_plural("(x|ch|ss|sh)$", "$1es")
	add_plural("([^aeiouy]|qu)ies$", "$1y")
	add_plural("([^aeiouy]|qu)y$", "$1ies")
	add_plural("(hive)$", "$1s")
	add_plural("(?:([^f])fe|([lr])f)$", "$1$2ves")
	add_plural("sis$", "ses")
	add_plural("([ti])um$", "$1a")
	add_plural("(buffal|tomat)o$", "$1oes")
	add_plural("(bu)s$", "$1es")
	add_plural("(alias|status)$", "$1es")
	add_plural("(octop|vir)us$", "$1i")
	add_plural("(ax|test)is$", "$1es")
	add_plural("s$", "s")
	add_plural("$", "s")

	add_singular("(quiz)zes$", "$1")
	add_singular("(matr)ices$", "$1ix")
	add_singular("(vert|ind)ices$", "$1ex")
	add_singular("^(ox)en", "$1")
	add_singular("(alias|status)es$", "$1")
	add_singular("([octop|vir])i$", "$1us")
	add_singular("(cris|ax|test)es$", "$1is")
	add_singular("(shoe)s$", "$1")
	add_singular("(o)es$", "$1")
	add_singular("(bus)es$", "$1")
	add_singular("([m|l])ice$", "$1ouse")
	add_singular("(x|ch|ss|sh)es$", "$1")
	add_singular("(m)ovies$", "$1ovie")
	add_singular("(s)eries$", "$1eries")
	add_singular("([^aeiouy]|qu)ies$", "$1y")
	add_singular("([lr])ves$", "$1f")
	add_singular("(tive)s$", "$1")
	add_singular("(hive)s$", "$1")
	add_singular("([^f])ves$", "$1fe")
	add_singular("(^analy)ses$", "$1sis")
	add_singular("((a)naly|(b)a|(d)iagno|(p)arenthe|(p)rogno|(s)ynop|(t)he)ses$", "$1$2sis")
	add_singular("([ti])a$", "$1um")
	add_singular("(n)ews$", "$1ews")
	add_singular("s$", "")

	uncountable("equipment", "information", "rice", "money", "species", "series", "fish", "sheep")
}

// check str only contains [a-zA-Z] or [.]
func IsCharDot(str string) bool {
	if "" == str {
		return false
	}
	for _, ch := range str {
		if (ch < 65 || ch > 122 || (ch > 90 && ch < 97)) && (ch != '.') {
			return false
		}
	}
	return true
}

// check str only contains [a-zA-Z] or [-_]or[0-9]
func IsCharDigitUnderscoreHyphen(str string) bool {
	if "" == str {
		return false
	}

	for _, ch := range str {
		if (ch < 65 || ch > 122 || (ch > 90 && ch < 97)) &&
			(ch != '_') &&
			(ch != '-') &&
			!unicode.IsDigit(ch) {
			return false
		}
	}
	return true
}

/**
 * Sample:
 * <p><b> beta_soft => BetaSoft </b></p>
 *
 * @param name the string formed in underscore
 * @return camel string
 */
func CamelCase(name string) string {
	newstr := make([]rune, 0, len(name))
	upNextChar := true

	for _, chr := range name {
		switch {
		case upNextChar:
			upNextChar = false
			chr -= ('a' - 'A')
		case chr == '_':
			upNextChar = true
			continue
		}

		newstr = append(newstr, chr)
	}

	return string(newstr)
}

/**
 * Underscore a word, such as:
 * <p/>
 * <p><b>BetaSoft -> beta_soft</b></p>
 *
 * @param camelCasedWord the camel case word
 * @return the underscored word
 */
func Underscore(name string) string {
	newstr := make([]rune, 0, len(name))
	firstTime := true

	for _, chr := range name {
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			if firstTime == true {
				firstTime = false
			} else {
				newstr = append(newstr, '_')
			}
			chr -= ('A' - 'a')
		}
		newstr = append(newstr, chr)
	}

	return string(newstr)
}

func Pluralize(str string) string {
	l := strings.ToLower(str)
	for _, s := range uncountables {
		if s == l {
			return str
		}
	}

	if strings.HasSuffix(str, "y") {
		str = str[:len(str)-1] + "ie"
	}
	return str + "s"
}

/**
 * Register two string are not regular, such as:
 * <p/>
 * person -> people
 *
 * @param add_singular add_singular form string
 * @param add_plural   add_plural form string
 */
func irregular(singular, plural string) {
	add_plural(singular+"$", plural)
	add_singular(plural+"$", singular)
}

/**
 * Register word add_plural rule
 *
 * @param rule        the rule(regexp string value)
 * @param replacement the replacement for previous regexp captured group
 */
func add_plural(rule, replacement string) {
	plurals = append(plurals, &RuleAndReplacement{rule: rule, replacement: replacement})
}

/**
 * Register word add_singular rule
 *
 * @param rule        the rule(regexp string value)
 * @param replacement the replacement for previous regexp captured group
 */
func add_singular(rule, replacement string) {
	singulars = append(singulars, &RuleAndReplacement{rule: rule, replacement: replacement})
}

/**
 * Register uncountable words
 *
 * @param words the uncountable words
 */
func uncountable(words ...string) {
	for _, s := range words {
		uncountables = append(uncountables, s)
	}
}

/**
 * Return word's add_plural form value
 *
 * @param word the word
 * @return the add_plural word
 */
func pluralizeFull(word string) string {
	l := strings.ToLower(word)
	for _, s := range uncountables {
		if s == l {
			return word
		}
	}
	return replaceWithFirstRule(l, plurals)
}

/**
 * Return the word's add_singular form value
 *
 * @param word the word
 * @return the add_singular word
 */
func singularize(word string) string {
	l := strings.ToLower(word)
	for _, s := range uncountables {
		if s == l {
			return word
		}
	}

	return replaceWithFirstRule(l, singulars)
}

/**
 * Return a tableized word for class name, such as:
 * <p><b>BetaSoft -> beta_softs </b></p>
 *
 * @param className the class name
 * @return the underscored and add_plural words
 */
func Tableize(className string) string {
	return Pluralize(Underscore(className))
}

func replaceWithFirstRule(word string, ruleAndReplacements []*RuleAndReplacement) string {
	// for _,RuleAndReplacement rar := range ruleAndReplacements {
	//     rule := rar.rule;
	//     replacement := rar.replacement

	//     matcher := syntax.Compile(rule, syntax.FoldCase)
	//             .matcher(word);
	//     if (matcher.find()) {
	//         return matcher.replaceAll(replacement);
	//     }
	// }
	return word
}
