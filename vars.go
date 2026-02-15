package doze

import (
	"bytes"
	"fmt"
	"text/template"
)

type Vars map[string]any

// Take a pointer to a string or list of strings and interpolate Vars into the string(s).
func Interpolate(dest any, vars *Vars) error {
	switch v := dest.(type) {
	case *string:
		template, err := template.New("VarsTemplate").Parse(*v)
		if err != nil {
			return fmt.Errorf("failed to create VarsTemplate from `%v`: %v", *v, err)
		}
		var buf bytes.Buffer
		if err := template.Execute(&buf, vars); err != nil {
			return fmt.Errorf("failed to execute VarsTemplate with `%v`: %v", *v, err)
		}
		*v = buf.String()
		return nil
	case *[]string:
		for i := range *v {
			if err := Interpolate(&(*v)[i], vars); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported type %T", dest)
	}
}
