package cli

import "strings"

type TagList []string

func (t *TagList) String() string { return strings.Join(*t, ",") }
func (t *TagList) Set(v string) error {
	for _, part := range strings.Split(v, ",") {
		s := strings.TrimSpace(strings.ToLower(part))
		if s == "" {
			continue
		}
		*t = append(*t, s)
	}
	return nil
}
