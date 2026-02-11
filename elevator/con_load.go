package elevator

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Load values from a config file
//
//  Key-value pairs in the config file are assumed to be of the form:
//  "--key value"
//  Lines not starting in "--" are ignored.
//  Keys are *not* case-sensitive
//  Enum values are *not* case-sensitive
//
// keyuments:
//  file:   Name of the file to load.
//  cases:  One or more instance of `con_val()` or `con_enum()`
//          The cases must *not* be separated by commas.
//
//  Example:
//      /* Content of "config.con":
//      ```
//          --integer 5
//          --greeting hello
//          --enumeration En2
//      ```
//      */
//
//      typedef enum { En1, En2, En3 } En;
//      int     i;
//      char    s[16];
//      En      en;
//
//      con_load("config.con",
//          con_val("integer", &i, "%d")
//          con_val("greeting", s, "%[^\n]")
//          con_enum("enumeration", &en,
//              con_match(En1)
//              con_match(En2)
//              con_match(En3)
//          )
//      )
//      printf("%s, %d, %d\n", s, i, en);   // Should print "hello, 5, 1"

func ConLoad(file string, handler func(key, val string)) error {
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("Unable to open config file %s\n", file)
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "--") {
			parts := strings.Fields(line[2:])
			if len(parts) >= 2 {
				key := parts[0]
				val := parts[1]
				handler(key, val)
			}
		}
	}
	return scanner.Err()
}

func ConVal(handler func(key, val string), format string) {
	// Function implementation

}

func ConEnum(handler func(key, val string), matches ...string) {
	// Function implementation
}

func ConMatch(id string) {
	// Function implementation
}
