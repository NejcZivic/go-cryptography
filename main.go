package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"
)

func stringToBinary(s string) string {
	var binaryStr strings.Builder

	for _, char := range s {
		binary := fmt.Sprintf("%08b", char)
		binaryStr.WriteString(binary)
	}

	return binaryStr.String()
}

func binaryToString(binary string) (string, error) {
	var result string
	for len(binary)%8 != 0 {
		binary = "0" + binary
	}

	for i := 0; i < len(binary); i += 8 {
		byteStr := binary[i : i+8]

		value, err := strconv.ParseInt(byteStr, 2, 8)
		if err != nil {
			return "", err
		}

		result += string(rune(value))
	}

	return result, nil
}

func main() {
	separator := strings.Repeat("-", 40)
	var input string
	consoleReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\033[H\033[2J")
		fmt.Printf("GO CRYPTOGRAPHY\n%s\n1. Encrypt text\n2. Decrypt image\n3. Exit\n%s\n", separator, separator)
		fmt.Scanln(&input)

		if input == "1" {
			var inputFile string
			var outputFile string
			var text string
			var readerErr error
			isJpeg := false

			fmt.Print("\033[H\033[2J")
			fmt.Print("Input file (with extension): ")
			fmt.Scanln(&inputFile)
			fmt.Print("Output file (with extension): ")
			fmt.Scanln(&outputFile)
			fmt.Print("The text to encode: ")
			text, readerErr = consoleReader.ReadString('\n')
			if readerErr != nil {
				panic(readerErr)
			}

			if strings.Split(inputFile, ".")[1] == "jpg" || strings.Split(inputFile, ".")[1] == "jpeg" {
				isJpeg = true
			}

			reader, openErr := os.Open(inputFile)
			if openErr != nil {
				panic(openErr)
			}

			var imageR image.Image
			var err error

			if isJpeg {
				imageR, err = jpeg.Decode(reader)
				if err != nil {
					panic(err)
				}
			} else {
				imageR, err = png.Decode(reader)
				if err != nil {
					panic(err)
				}
			}

			reader.Close()

			bounds := imageR.Bounds()
			width, height := bounds.Max.X, bounds.Max.Y
			img := image.NewRGBA(bounds)

			draw.Draw(img, bounds, imageR, image.Point{}, draw.Over)

			binary := stringToBinary(text)
			stamp := fmt.Sprintf("%16s", strconv.FormatInt(int64(len(binary)), 2))
			stamp = strings.ReplaceAll(stamp, " ", "0")
			binary = stamp + binary

			counter := 0

			for y := range height {
				for x := range width {
					pixelColor := img.At(x, y)

					r, g, b, a := pixelColor.RGBA()

					r8 := uint8(r >> 8)

					if counter < len(binary) {
						if (string(binary[counter])) == "0" {
							r8 &= ^uint8(1)
						} else {
							r8 |= 1
						}

						counter++
					} else {
						break
					}

					rgba := color.NRGBA{r8, uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}

					img.Set(x, y, rgba)
				}
			}

			file, createErr := os.Create(outputFile)
			if createErr != nil {
				panic(createErr)
			}

			var writer io.Writer = file

			if isJpeg {
				jpeg.Encode(writer, img, nil)
			} else {
				png.Encode(writer, img)
			}

			file.Close()
			continue
		}

		if input == "2" {
			var file string
			isJpeg := false

			fmt.Print("\033[H\033[2J")
			fmt.Print("File to decrypt (with extension): ")
			fmt.Scanln(&file)

			if strings.Split(file, ".")[1] == "jpg" || strings.Split(file, ".")[1] == "jpeg" {
				isJpeg = true
			}

			reader, openErr := os.Open(file)
			if openErr != nil {
				panic(openErr)
			}

			var imageR image.Image
			var err error

			if isJpeg {
				imageR, err = jpeg.Decode(reader)
				if err != nil {
					panic(err)
				}
			} else {
				imageR, err = png.Decode(reader)
				if err != nil {
					panic(err)
				}
			}

			bounds := imageR.Bounds()
			width, height := bounds.Max.X, bounds.Max.Y

			lenBinary := ""
			lenCounter := 0

			decryptedBinary := ""
			decryptionCounter := 0
			startDecrypting := false
			var length int64

			failed := false

			for y := range height {
				if failed {
					break
				}
				for x := range width {
					pixelColor := imageR.At(x, y)

					r, _, _, _ := pixelColor.RGBA()

					r8 := uint8(r >> 8)

					if !startDecrypting {
						if lenCounter != 16 {
							lsb := r8 & 1
							lenBinary += strconv.Itoa(int(lsb))
							lenCounter++
							continue
						} else {
							length, err = strconv.ParseInt(lenBinary, 2, 64)
							if err != nil {
								failed = true
								fmt.Println("There is nothing encoded.")
								fmt.Scanln()
							}
							startDecrypting = true
						}
					}

					if decryptionCounter < int(length) {
						lsb := r8 & 1
						decryptedBinary += strconv.Itoa(int(lsb))
						decryptionCounter++
					} else {
						break
					}
				}
			}

			if failed {
				continue
			}

			msg, conversionErr := binaryToString(decryptedBinary)
			if conversionErr != nil {
				panic(conversionErr)
			}

			fmt.Printf("Decrypted message: %s\n", msg)
			fmt.Scanln()
			continue
		}

		if input == "3" {
			os.Exit(0)
		}

		fmt.Println("This isn't a valid option")
		fmt.Scanln()
		continue
	}
}
