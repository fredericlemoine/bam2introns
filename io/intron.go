package io

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/biogo/hts/sam"
)

type Intron struct {
	Chr    string // chromosome
	Start  int    // [ : inclusive / 0 based
	End    int    // [ : exclusive / 1 based
	Strand bool   // true : + , false: -
	Mate   bool   // true : from a mate read, false : from a first read
	Name   string // Name of the read
	Count  []int  // May store several read counts (if several samples for example), nil by default
}

type Stranded byte

const (
	NONE     Stranded = iota // Not stranded
	STRANDED                 // Stranded Library +/-
	REVERSE                  // Reverse stranded library -/+
)

func Strand(s string) Stranded {
	switch s {
	case "none":
		return NONE
	case "stranded":
		return STRANDED
	case "reverse":
		return REVERSE
	default:
		ExitWithMessage(errors.New("\"" + s + "\": is not a valid stranded type "))
	}
	return 0
}

/*
Returns the introns represented by this read (using CIGAR N operator)
If any
*/
func Introns(read *sam.Record, s Stranded) []*Intron {
	introns := make([]*Intron, 0, 3)
	start := read.Start()

	strand := read.Strand() == 1
	mate := read.Flags&sam.Read2 != 0
	switch s {
	case STRANDED:
		if read.Flags&sam.Read2 != 0 {
			strand = !strand
		}
	case REVERSE:
		if read.Flags&sam.Read1 != 0 {
			strand = !strand
		}
	}

	for _, cigarOp := range read.Cigar {
		switch cigarOp.Type() {
		case sam.CigarSkipped:
			introns = append(introns, &Intron{read.Ref.Name(), start, start + cigarOp.Len(), strand, mate, read.Name, nil})
			start += cigarOp.Len()
		case sam.CigarMatch:
			start += cigarOp.Len()
		case sam.CigarInsertion:
			// Nothing to add to the start
		case sam.CigarDeletion:
			start += cigarOp.Len()
		case sam.CigarSoftClipped:
			// Nothing to add to the start
		case sam.CigarHardClipped:
			// Nothing to add to the start
		case sam.CigarPadded:
			// ?
		case sam.CigarEqual:
			start += cigarOp.Len()
		case sam.CigarMismatch:
			start += cigarOp.Len()
		case sam.CigarBack:
			// ?
		}
	}
	return introns
}

func PrintIntrons(i *Intron) string {
	if i.Strand {
		return fmt.Sprintf("%s\t%d\t%d\t%s\t%s\t+", i.Chr, i.Start, i.End, i.Name, joinCounts(i))
	} else {
		return fmt.Sprintf("%s\t%d\t%d\t%s\t%s\t-", i.Chr, i.Start, i.End, i.Name, joinCounts(i))
	}
}

// Returns string representation of read counts for the intron
// if no counts (nil): then "0"
func joinCounts(i *Intron) string {

	if i.Count == nil {
		return "0"
	}

	var buffer bytes.Buffer
	for id, c := range i.Count {
		if id > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(fmt.Sprintf("%d", c))
	}
	return buffer.String()
}
