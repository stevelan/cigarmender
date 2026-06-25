package bamreader

import (
	"cigarmender/reference"
	"fmt"
	"log/slog"

	"github.com/biogo/hts/sam"
)

// NewDelCentrer creates a DelCenterer
func NewDelCentrer() *DelCentrer {
	return &DelCentrer{}
}

type DelCentrer struct {
	DelCount int
	HPCount  int
	Rewrites int
}

func (d DelCentrer) Summary() string {
	return fmt.Sprintf("Found %d homopolymers out of %d deletions, with %d rewrites", d.HPCount, d.DelCount, d.Rewrites)
}

/**
* Counts the number of reads with deletions
 */
func (d *DelCentrer) Visit(read *sam.Record, hpIndex *reference.RefIndex, bamWriter *BamWriter) error {

	newCigar, modified := processCigar(read, hpIndex, d)

	if modified {
		slog.Debug("Writing new CIGAR for read", "read", read.Name, "cigar", newCigar)
		return bamWriter.WriteToBam(read, newCigar)
	} else {
		slog.Debug("Writing existing unmodified read", "read", read.Name, "cigar", newCigar)
		return bamWriter.WriteToBamExisting(read)
	}
}

func processCigar(read *sam.Record, hpIndex *reference.RefIndex, d *DelCentrer) ([]sam.CigarOp, bool) {
	var (
		pendingCigarDel sam.CigarOp
		pendingHpRange  reference.Range
		hasPending      bool
		pendingDelRPos  int
	)

	// start at read position
	rpos := read.Pos

	newCigar := make([]sam.CigarOp, 0, len(read.Cigar))
	any_modified := false

	for _, cigarop := range read.Cigar {
		if cigarop.Type() == sam.CigarDeletion && !hasPending {

			// check if hp using reference co-ordinates
			query := reference.NewRange(rpos, rpos+cigarop.Len())
			hp, found := hpIndex.Search(read.Ref.Name(), query)
			d.DelCount++
			if found && len(newCigar) > 0 {
				slog.Debug("Found homopolymer for read", "read", read.Name, "hp", hp.String())
				pendingCigarDel = cigarop
				pendingHpRange = hp
				hasPending = true
				pendingDelRPos = rpos
				d.HPCount++
			} else {
				// not in homopolymer or this is the first cigarop, just append to CIGAR
				hasPending = false
				newCigar = append(newCigar, cigarop)
			}
		} else {
			// if the op prior to the deletion was a match and this op is a match, we can recentre the deletion
			if hasPending {
				slog.Debug("Has pending deletion", "delStart", pendingDelRPos, "range", pendingHpRange)
				lastPushedOp := newCigar[len(newCigar)-1]
				// need the last pushed op and the current op to be a match to rewrite the deletion
				if isMatch(lastPushedOp) && isMatch(cigarop) {
					slog.Debug("Last pushed op and current op are match", "lastPushedOp", lastPushedOp.Type(), "current", cigarop.Type())
					poppedCigar := newCigar[:len(newCigar)-1]
					// rewrite the cigar
					cigarFragment, modified := rewriteCigar(lastPushedOp, pendingCigarDel, cigarop, pendingHpRange, pendingDelRPos)
					newCigar = append(poppedCigar, cigarFragment...)
					any_modified = any_modified || modified
					d.Rewrites++
				} else {
					slog.Debug("Last pushed op and current op are not a match", "lastPushedOp", lastPushedOp.Type(), "current", cigarop.Type())
					// just push the deletion and current op onto the stack
					newCigar = append(newCigar, pendingCigarDel, cigarop)
				}
			} else {
				newCigar = append(newCigar, cigarop)
			}
			hasPending = false
			pendingCigarDel = 0
			pendingHpRange = reference.NewRange(0, 0)
		}
		// increment reference if the cigarop consumes it
		rpos += cigarop.Len() * cigarop.Type().Consumes().Reference
	}

	if hasPending {
		// push the last pendingDel, no subsequent match to centre within
		newCigar = append(newCigar, pendingCigarDel)
	}
	return newCigar, any_modified
}

func isMatch(cigarop sam.CigarOp) bool {
	optype := cigarop.Type()
	return optype == sam.CigarMatch ||
		optype == sam.CigarEqual
}

func rewriteCigar(
	priorMatch sam.CigarOp,
	deletion sam.CigarOp,
	nextMatch sam.CigarOp,
	homopolymer reference.Range,
	delRpos int,
) (cigar []sam.CigarOp, modified bool) {
	delLen := deletion.Len()
	hpLen := homopolymer.Len()

	if delLen >= hpLen {
		slog.Debug("Deletion len is greater than hp length", "delLen", delLen, "hpLen", hpLen)
		return []sam.CigarOp{priorMatch, deletion, nextMatch}, false
	}

	// Centre deletion within the homopolymer.
	// For a homopolymer [start, end), the latest valid deletion start is end - delLen.
	targetDelStart := homopolymer.Start + (hpLen-delLen)/2

	shift := targetDelStart - delRpos

	if shift == 0 {
		slog.Debug("Shift is zero", "prior", priorMatch, "deletion", deletion, "next", nextMatch, "hpStart", homopolymer.Start, "hpLen", hpLen, "delLen", delLen)
		return []sam.CigarOp{priorMatch, deletion, nextMatch}, false
	}

	// Positive shift moves the deletion to the right:  prior match gets longer, next match gets shorter.
	//
	// Negative shift moves the deletion to the left:  prior match gets shorter, next match gets longer.
	newPriorLen := priorMatch.Len() + shift
	newNextLen := nextMatch.Len() - shift

	for newPriorLen <= 0 {
		newPriorLen++
		newNextLen--
	}

	for newNextLen <= 0 {
		newNextLen++
		newPriorLen--
	}

	if newPriorLen < 0 || newNextLen < 0 {
		// Cannot realise this rewrite with only the immediate flanking ops.
		slog.Warn("Prior or next are less than zero", "prior", newPriorLen, "next", newNextLen)
		return []sam.CigarOp{priorMatch, deletion, nextMatch}, false
	}

	cleanedCigar := cleanCigarOps([]sam.CigarOp{
		sam.NewCigarOp(priorMatch.Type(), newPriorLen),
		deletion,
		sam.NewCigarOp(nextMatch.Type(), newNextLen),
	})
	slog.Debug("Rewriting cigar successfully", "prior", priorMatch, "deletion", deletion, "next", nextMatch, "new", cleanedCigar)
	return cleanedCigar, true
}

// removes zero length cigar ops and merges adjacent ops of the same type
func cleanCigarOps(ops []sam.CigarOp) []sam.CigarOp {
	out := make([]sam.CigarOp, 0, len(ops))

	for _, op := range ops {
		if op.Len() == 0 {
			continue
		}

		if len(out) > 0 && out[len(out)-1].Type() == op.Type() {
			prev := out[len(out)-1]
			out[len(out)-1] = sam.NewCigarOp(prev.Type(), prev.Len()+op.Len())
			continue
		}

		out = append(out, op)
	}

	return out
}
