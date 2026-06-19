# CIGARMender

CIGARMender - homopolymer-aware deletion centering of aligned reads that preserves modification tags

Homopolymer slippage in nanopore sequencing leads to errors in alignments, these are often left aligned by aligners
to the start of the homopolymer runs which can split information rich k-mers. This tool centres deletions within 
homopolymer runs so that motifs at either end of the homopolymer run are not arbitrarily split.

Currently: 

|Read| Seq            | Cigar   | Mod motif | 
| ---| ---------------| --------| ----------|
| R1 | GAAAAa------CT | M6D6M2  | AAaAA     |
| R2 | GAAAAAAa----CT | M8D4M2  | AAaAA     |
| R3 | GAAAAAAAAAa-CT | M11D1M2 | AAaAC     |
| Ref| GAAAAAAAAAAACT |         |           |

With CIGARMender: 

|Read| Seq            | Cigar   | Mod motif |
| ---| ---------------| --------| ----------|
| R1 | GAA------AAaCT | M3D6M5  | AAaCT     |
| R2 | GAAA----AAAaCT | M4D4M6  | AAaCT     |
| R3 | GAAAAA-AAAAaCT | M6D1M7  | AAaCT     |
| Ref| GAAAAAAAAAAACT |         |           |


## Getting started
Install required tools with \`make wintools`
Build the project with `make build` will produce a binary in the root of the project directory
Clean the project with `make clean` 
