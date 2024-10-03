package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"
)

var opcode uint16     // current opcode
var memory [4096]byte // 4K memory
var V [16]byte        // 16 8-bit registers

var I uint16  // index register
var pc uint16 // program counter

/*
0x000-0x1FF - Chip 8 interpreter (contains font set in emu)
0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
0x200-0xFFF - Program ROM and work RAM
*/

const width int = 64
const height int = 32

var gfx [width * height]byte // 2048 pixels
var delayTimer byte
var soundTimer byte

var stack [16]uint16 // stack with 16 levels
var sp byte          // stack pointer

var key [16]byte // 16 keys
var drawFlag bool

var program_add string = "pong2.c8"

var fontset = [80]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

func Initialize() {

	rand.Seed(time.Now().UnixNano())

	pc = 0x200
	opcode = 0
	I = 0
	sp = 0

	// clear display
	for i := 0; i < 2048; i++ {
		gfx[i] = 0
	}
	// clear stack
	for i := 0; i < 16; i++ {
		stack[i] = 0
	}
	// clear registers V0-VF
	for i := 0; i < 16; i++ {
		key[i] = 0
		V[i] = 0
	}
	// clear memory
	for i := 0; i < 4096; i++ {
		memory[i] = 0
	}

	for i := 0; i < 80; i++ {
		memory[i] = fontset[i]
	}

	// read from rom
	file, err := os.Open(program_add)
	if err != nil {
		panic(err)
	}
	buffer, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	bufferSize := len(buffer)
	fmt.Println(bufferSize)
	for i := 0; i < bufferSize; i++ {
		memory[i+512] = buffer[i]
	}

	// reset timers
	delayTimer = 0
	soundTimer = 0

	drawFlag = true
}

func EmulateCycle() {
	opcode = uint16(memory[pc])<<8 | uint16(memory[pc+1])
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode & 0x000F {
		case 0x0000:
			// clear the screen
			for i := 0; i < 2048; i++ {
				gfx[i] = 0
			}
			pc += 2
		case 0x000E:
			// return from subroutine
			sp--
			pc = stack[sp]
			pc += 2
		default:
			fmt.Printf("Unknown opcode: 0x%X\n", opcode)
		}
	case 0x1000: // 1NNN: Jump to address NNN
		pc = opcode & 0x0FFF
	case 0x2000:
		stack[sp] = pc
		sp++
		pc = opcode & 0x0FFF
	case 0x3000: // 3XNN: Skip next instruction if VX == NN
		if V[(opcode&0x0F00)>>8] == byte(opcode&0x00FF) {
			pc += 4
		} else {
			pc += 2
		}
	case 0x4000: // 4XNN: Skip next instruction if VX != NN
		if V[(opcode&0x0F00)>>8] != byte(opcode&0x00FF) {
			pc += 4
		} else {
			pc += 2
		}
	case 0x5000: // 5XY0: Skip next instruction if VX == VY
		if V[(opcode&0x0F00)>>8] == V[(opcode&0x00F0)>>4] {
			pc += 4
		} else {
			pc += 2
		}
	case 0x6000: // 6XNN: Set VX to NN
		V[(opcode&0x0F00)>>8] = byte(opcode & 0x00FF)
		pc += 2

	case 0x7000: // 7XNN: Add NN to VX
		V[(opcode&0x0F00)>>8] += byte(opcode & 0x00FF)
		pc += 2

	case 0x8000:
		switch opcode & 0x000F {
		case 0x0000: // 8XY0: Set VX to the value of VY
			V[(opcode&0x0F00)>>8] = V[(opcode&0x00F0)>>4]
			pc += 2
		case 0x0001: // 8XY1: Set VX to VX OR VY
			V[(opcode&0x0F00)>>8] |= V[(opcode&0x00F0)>>4]
			pc += 2
		case 0x0002: // 8XY2: Set VX to VX AND VY
			V[(opcode&0x0F00)>>8] &= V[(opcode&0x00F0)>>4]
			pc += 2

		case 0x0003: // 8XY3: Set VX to VX XOR VY
			V[(opcode&0x0F00)>>8] ^= V[(opcode&0x00F0)>>4]
			pc += 2
		case 0x0004: // 8XY4: Add VY to VX.
			//	VF is set to 1 when there's a carry
			if V[(opcode&0x0F00)>>8] > 0xFF-V[(opcode&0x00F0)>>4] {
				V[0xF] = 1 // carry
			} else {
				V[0xF] = 0
			}
			V[(opcode&0x0F00)>>8] += V[(opcode&0x00F0)>>4]
			pc += 2
		case 0x0005: // 8XY5: Subtract VY from VX
			// VF is set to 0 when there's a borrow
			if V[(opcode&0x0F00)>>4] > V[(opcode&0x00F0)>>8] {
				V[0xF] = 1 // there is a borrow
			} else {
				V[0xF] = 0
			}
			V[(opcode&0x0F00)>>8] -= V[(opcode&0x00F0)>>4]
			pc += 2
		case 0x0006: // 8XY6: Shift VX right by 1
			V[0xF] = V[(opcode&0x0F00)>>8] & 0x1
			V[(opcode&0x0F00)>>8] >>= 1
			pc += 2

		case 0x0007: // 8XY7: Set VX to VY minus VX
			// VF is set to 0 when there's a borrow
			if V[(opcode&0x0F00)>>8] > V[(opcode&0x00F0)>>4] {
				V[0xF] = 0 // there is a borrow
			} else {
				V[0xF] = 1
			}
			V[(opcode&0x0F00)>>8] = V[(opcode&0x00F0)>>4] - V[(opcode&0x0F00)>>8]
			pc += 2
		case 0x000E: // 8XYE: Shift VX left by 1
			V[0xF] = V[(opcode&0x0F00)>>8] >> 7
			V[(opcode&0x0F00)>>8] <<= 1
			pc += 2

		default:
			fmt.Printf("Unknown opcode: 0x%X\n", opcode)
		}
	case 0x9000: // 9XY0: Skip next instruction if VX != VY
		if V[(opcode&0x0F00)>>8] != V[(opcode&0x00F0)>>4] {
			pc += 4
		} else {
			pc += 2
		}

	case 0xA000:
		I = opcode & 0x0FFF
		pc += 2

	case 0xB000:
		pc = (opcode & 0x0FFF) + uint16(V[0])

	case 0xC000: // CXNN: Set VX to a random number and NN
		V[(opcode&0x0F00)>>8] = byte((uint16(rand.Int()) & 0x00FF) & uint16(opcode&0x00FF))
		log.Printf("Random number: %X", V[(opcode&0x0F00)>>8])
		pc += 2

	case 0xD000:
		x := int(V[(opcode&0x0F00)>>8])
		y := int(V[(opcode&0x00F0)>>4])
		h := opcode & 0x000F
		var pixel byte

		V[0xF] = 0
		for yline := uint16(0); yline < h; yline++ {
			pixel = memory[I+yline]
			for xline := uint16(0); xline < 8; xline++ {
				if x+int(xline) == width {
					x = -int(xline)
				}
				if y+int(yline) == height {
					y = -int(yline)
				}
				if (pixel & (0x80 >> xline)) != 0 {
					if gfx[(int(x)+int(xline)+((int(y)+int(yline))*64))] == 1 {
						V[0xF] = 1
					}
					gfx[int(x)+int(xline)+((int(y)+int(yline))*64)] ^= 1
				}
			}
		}
		drawFlag = true
		pc += 2

	case 0xE000:
		switch opcode & 0x00FF {
		case 0x009E:
			if key[V[(opcode&0x0F00)>>8]] != 0 {
				pc += 4
			} else {
				pc += 2
			}
		case 0x00A1:
			if key[V[(opcode&0x0F00)>>8]] == 0 {
				pc += 4
			} else {
				pc += 2
			}
		default:
			fmt.Printf("Unknown opcode: 0x%X\n", opcode)
		}

	case 0xF000:
		switch opcode & 0x00FF {
		case 0x0007: // FX07: Set VX to the value of the delay timer
			V[(opcode&0x0F00)>>8] = delayTimer
			pc += 2
		case 0x000A: // FX0A: Wait for a key press, store the value of the key in VX
			keyPress := false
			for i := 0; i < 16; i++ {
				if key[i] != 0 {
					V[(opcode&0x0F00)>>8] = byte(i)
					keyPress = true
				}
			}
			if !keyPress {
				return
			}
			pc += 2
		case 0x0015: // FX15: Set the delay timer to VX
			delayTimer = V[(opcode&0x0F00)>>8]
			pc += 2

		case 0x0018: // FX18: Set the sound timer to VX
			soundTimer = V[(opcode&0x0F00)>>8]
			pc += 2
		case 0x001E: // FX1E: Add VX to I
			if I+uint16(V[(opcode&0x0F00)>>8]) > 0xFFF {
				V[0xF] = 1
			} else {
				V[0xF] = 0
			}
			I += uint16(V[(opcode&0x0F00)>>8])
			pc += 2

		case 0x0029:
			I = uint16(V[(opcode&0x0F00)>>8]) * 0x5
			pc += 2

		case 0x0033:
			memory[I] = V[(opcode&0x0F00)>>8] / 100
			memory[I+1] = (V[(opcode&0x0F00)>>8] / 10) % 10
			memory[I+2] = (V[(opcode&0x0F00)>>8] % 100) % 10
			pc += 2
		case 0x0055: // FX55: Store registers V0 through VX in memory starting at location I
			for i := uint16(0); i <= (opcode&0x0F00)>>8; i++ {
				memory[I+i] = V[i]
			}
			I += ((opcode & 0x0F00) >> 8) + 1
			pc += 2

		case 0x0065:
			for i := uint16(0); i <= (opcode&0x0F00)>>8; i++ {
				V[i] = memory[I+i]
			}
			I += ((opcode & 0x0F00) >> 8) + 1
			pc += 2

		default:
			fmt.Printf("Unknown opcode: 0x%X\n", opcode)
		}

	default:
		fmt.Printf("Unknown opcode: 0x%X\n", opcode)
	}

	// Update timers
	if delayTimer > 0 {
		delayTimer--
	}
	if soundTimer > 0 {
		if soundTimer == 1 {
			fmt.Println("BEEP!")
		}
		soundTimer--
	}
}
