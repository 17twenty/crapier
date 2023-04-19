package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	{

		f, err := os.Create("response.html")

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		_, err2 := f.WriteString(`
<!DOCTYPE html>
<html>
	<head>
		<title>Example</title>
	</head>
	<body>
		<p>Hello World</p>
	</body>
</html>
	`)

		if err2 != nil {
			log.Fatal(err2)
		}
	}

	{

		f, err := os.Create("headers.out")

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		_, err2 := f.WriteString(`Content-Type: text/html; charset=utf-8`)

		if err2 != nil {
			log.Fatal(err2)
		}
	}

	fmt.Println("done")
}
