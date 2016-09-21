# bam2introns

Extracts intron coordinates from a bam file, and writes them in bed format.

Takes into account orientation of the reads, depending on option -s.

May group the introns per coordinates and count the number of reads (option -g).

Example:
```bash
bam2introns -s none -t 2  sample.bam  > sample_introns.bed
```
