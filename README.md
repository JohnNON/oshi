# oshi

oshi is an [oshi.at](https://oshi.at) api client.

Installation

    go get github.com/JohnNON/oshi

Example of usage:

```golang
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/JohnNON/oshi"
)

func main() {
	ctx := context.Background()

	file, err := os.ReadFile("example.png")
	if err != nil {
		log.Fatalln(err)
	}

	client := oshi.NewClient(&http.Client{})

	img := oshi.NewImage(file, "name_test", 5, true, false, false)

	resp, err := client.Upload(ctx, img)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", resp)
}
```