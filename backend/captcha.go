package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var (
	captchas   = make(map[string]*Captcha)
	nonces     = make(map[string]time.Time)
	rateLimits = make(map[string]*RateLimit)
	mu         sync.Mutex

	captchaWords []string

	// AES-256-GCM encrypted captcha words
	encryptedCaptchaWords = []string{
		"0000000000000000000000006d68923d1bcc3b305eb79454237c4e8086aa94854ae542",
		"01010101010101010101010191e445b4a06dc2784bebe0882c9a2087a16df3b4133e08",
		"020202020202020202020202ecd6e677cf1af81a1a0a66cbcee5c569721a82c5b281",
		"030303030303030303030303756de00f7a62f5920ef0ed5ed9c5d8225d6375381967",
		"0404040404040404040404049650dac7e6298a52b91f276249a9fb2705d2e60ac3a5",
		"050505050505050505050505f2b3f46dd57a62c3233df0155696a09199bda790dff751",
		"060606060606060606060606541c40471143bf7355054240166920792b2a6b05dd8fc5",
		"070707070707070707070707036735427ba8dbdce2fe18889f047cb00ad8b9da5a2a",
		"0808080808080808080808080d08baa09fbacadd34280f46f02e479cda5eda20281d4c",
		"090909090909090909090909a6f40b599e19e607e1cdf7995d1691f0fc7cbb5a55e5",
		"0a0a0a0a0a0a0a0a0a0a0a0ad20834ebba97f2b5db63ad843e250c43e940a6f891ba",
		"0b0b0b0b0b0b0b0b0b0b0b0bd1f035839a0ae88a56eb1601986876ba0236ccff7516",
		"0c0c0c0c0c0c0c0c0c0c0c0c2dffaed40e4b03047df753602de2ceb9a39acaec2611",
		"0d0d0d0d0d0d0d0d0d0d0d0d80661169348a409de21effa3cbffa23d710e243a3e42d3",
		"0e0e0e0e0e0e0e0e0e0e0e0ea62b1d12f0c6f0919796ff869b86588f525eba5bba",
		"0f0f0f0f0f0f0f0f0f0f0f0f8257bc6a6d6ad315b1756240afb307bd308a978777c7",
		"1010101010101010101010109af865c3e4a977b2b34987aef93b86e55ba03e2794",
		"111111111111111111111111ede4a473dfd5accc2ad81e8d7ef758a3f22e97f2228e",
		"12121212121212121212121232c9b9f31eb45663692ef2e36d81e88116fff40a2feb",
		"1313131313131313131313134ea1d8fc7e206dce554cdd12e04cb75ca12fbfa649cd",
		"1414141414141414141414141db3ddc44a871fc064f3db95d3db5581a93761e731",
		"15151515151515151515151577be4a557c10ad69d53d53b0b7d29b5e1307ffc66418",
		"161616161616161616161616d3f789a9da64f6f40bc3403996e188668d62ab02f662",
		"171717171717171717171717dd4e0f41d941810d5d1765cef1b65339e4897a11cb3b",
		"181818181818181818181818fd59001ab8d33a2a3dcfd31382801f9bf38b2398b898",
		"1919191919191919191919193dea6ab30d4477acbacf057633b1d76ab2479e9b74",
		"1a1a1a1a1a1a1a1a1a1a1a1a7653dc05050f0508c5e9ae6d89187cfd3201665ba3",
		"1b1b1b1b1b1b1b1b1b1b1b1b935716c9f5b6f0930c8cf2e14b9575fec8b4906e6f5ca132",
		"1c1c1c1c1c1c1c1c1c1c1c1cc80bd3b04b30d0330804fead5553ecbe53e1bcff744c",
		"1d1d1d1d1d1d1d1d1d1d1d1d51b7fe60dbd7fe3ee139260f600a58f681f1d919cf",
	}

	captchaAESKey = []byte{
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
	}
)

func initCaptchaWords() {
	block, err := aes.NewCipher(captchaAESKey)
	if err != nil {
		log.Fatalf("Failed to create AES cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("Failed to create GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	captchaWords = make([]string, len(encryptedCaptchaWords))
	for i, enc := range encryptedCaptchaWords {
		data, err := hex.DecodeString(enc)
		if err != nil {
			log.Fatalf("Failed to decode captcha word %d: %v", i, err)
		}
		if len(data) < nonceSize {
			log.Fatalf("Encrypted captcha word %d is too short", i)
		}
		nonce, ciphertext := data[:nonceSize], data[nonceSize:]
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			log.Fatalf("Failed to decrypt captcha word %d: %v", i, err)
		}
		captchaWords[i] = string(plaintext)
	}
}

func randInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

func randWord() string {
	return captchaWords[randInt(len(captchaWords))]
}

func createCaptcha() *Captcha {
	id := uuid.New().String()
	answer := randWord()

	mu.Lock()
	captchas[id] = &Captcha{ID: id, Answer: answer, CreatedAt: time.Now()}
	mu.Unlock()

	return &Captcha{ID: id, Answer: answer, CreatedAt: time.Now()}
}

const (
	captchaTTL = 5 * time.Minute
	nonceTTL   = 10 * time.Minute
)

func consumeCaptcha(id string, nonceVal string) (*Captcha, bool) {
	mu.Lock()
	defer mu.Unlock()

	c, ok := captchas[id]
	if !ok || time.Since(c.CreatedAt) > captchaTTL {
		return nil, false
	}

	nonceCreated, known := nonces[nonceVal]
	if !known || time.Since(nonceCreated) > nonceTTL {
		return nil, false
	}

	delete(captchas, id)
	delete(nonces, nonceVal)
	return c, true
}

func captchaExists(id string) bool {
	mu.Lock()
	defer mu.Unlock()
	c, ok := captchas[id]
	if !ok || time.Since(c.CreatedAt) > captchaTTL {
		return false
	}
	return true
}

func drawCaptchaImage(answer string) []byte {
	width := 220
	height := 70

	bg := color.RGBA{R: 240, G: 240, B: 245, A: 255}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bg)
		}
	}

	for i := 0; i < 1500; i++ {
		x := randInt(width)
		y := randInt(height)
		c := color.RGBA{
			R: uint8(150 + randInt(105)),
			G: uint8(150 + randInt(105)),
			B: uint8(150 + randInt(105)),
			A: 255,
		}
		img.Set(x, y, c)
	}

	for i := 0; i < 6; i++ {
		x1 := randInt(width)
		y1 := randInt(height)
		x2 := randInt(width)
		y2 := randInt(height)
		lineColor := color.RGBA{
			R: uint8(100 + randInt(155)),
			G: uint8(100 + randInt(155)),
			B: uint8(100 + randInt(155)),
			A: 255,
		}
		drawLine(img, x1, y1, x2, y2, lineColor)
	}

	letterColors := []color.RGBA{
		{R: 30, G: 60, B: 150, A: 255},
		{R: 150, G: 30, B: 60, A: 255},
		{R: 30, G: 120, B: 60, A: 255},
		{R: 130, G: 70, B: 20, A: 255},
		{R: 90, G: 30, B: 130, A: 255},
	}

	face := basicfont.Face7x13
	totalWidth := len(answer) * 22
	startX := (width - totalWidth) / 2
	baseY := height/2 + 5

	for i, ch := range strings.ToUpper(answer) {
		dx := startX + i*22 + randInt(4) - 2
		dy := baseY + randInt(8) - 4
		c := letterColors[randInt(len(letterColors))]
		drawChar(img, dx, dy, ch, face, c)
	}

	for i := 0; i < 4; i++ {
		x1 := randInt(width)
		y1 := randInt(height)
		x2 := randInt(width)
		y2 := randInt(height)
		arcColor := color.RGBA{
			R: uint8(80 + randInt(150)),
			G: uint8(80 + randInt(150)),
			B: uint8(80 + randInt(150)),
			A: 255,
		}
		drawCurve(img, x1, y1, x2, y2, arcColor)
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func drawChar(img *image.RGBA, x, y int, ch rune, face font.Face, c color.Color) {
	dr, mask, maskp, advance, ok := face.Glyph(fixed.P(x, y), ch)
	if !ok {
		return
	}
	dr = dr.Intersect(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	if dr.Empty() {
		return
	}
	for dy := 0; dy < dr.Dy(); dy++ {
		for dx := 0; dx < dr.Dx(); dx++ {
			_, _, _, alpha := mask.At(maskp.X+dx, maskp.Y+dy).RGBA()
			if alpha > 0 {
				img.Set(dr.Min.X+dx, dr.Min.Y+dy, c)
			}
		}
	}
	_ = advance
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx - dy
	for {
		if x1 >= 0 && x1 < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
			img.Set(x1, y1, c)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func drawCurve(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	midx := (x1 + x2) / 2
	midy := (y1+y2)/2 - 15
	for t := 0.0; t <= 1.0; t += 0.01 {
		px := int((1-t)*(1-t)*float64(x1) + 2*(1-t)*t*float64(midx) + t*t*float64(x2))
		py := int((1-t)*(1-t)*float64(y1) + 2*(1-t)*t*float64(midy) + t*t*float64(y2))
		if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
			img.Set(px, py, c)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func handleCaptchaImage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/captcha/image/")
	id = sanitizeInput(id)

	if !captchaExists(id) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	mu.Lock()
	c := captchas[id]
	answer := c.Answer
	mu.Unlock()

	pngData := drawCaptchaImage(answer)

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Write(pngData)
}

func cleanupExpired() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for id, c := range captchas {
			if time.Since(c.CreatedAt) > captchaTTL {
				delete(captchas, id)
			}
		}
		now := time.Now()
		for n, created := range nonces {
			if now.Sub(created) > nonceTTL {
				delete(nonces, n)
			}
		}
		for ip, rl := range rateLimits {
			if time.Since(rl.LastTry) > 30*time.Minute {
				delete(rateLimits, ip)
			}
		}
		mu.Unlock()
		pruneExpiredTokens()
	}
}

func checkRateLimit(key string, maxAttempts int, window, lockout time.Duration) bool {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	rl, exists := rateLimits[key]
	if !exists {
		rateLimits[key] = &RateLimit{Attempts: 1, LastTry: now}
		return true
	}

	if now.Before(rl.LockedUntil) {
		return false
	}
	if now.Sub(rl.LastTry) > window {
		rl.Attempts = 0
	}
	rl.Attempts++
	rl.LastTry = now
	if rl.Attempts > maxAttempts {
		rl.LockedUntil = now.Add(lockout)
		return false
	}
	return true
}
