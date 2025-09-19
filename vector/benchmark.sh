#!/usr/bin/env bash

dir=benchmark
data=$dir/benchmark.dat

plot() {
	# Nth plot
	awk '/BenchmarkNth/{count ++; sub(/BenchmarkNth/,""); printf("%d,%s,%s,%s\n",count,$1,$2,$3)}' "$data" > "$dir/nth.dat"
	gnuplot \
		-e "file_path='$dir/nth.dat'" \
		-e "graphic_file_name='$dir/nth.png'" \
		-e "graphic_title='Nth()'" \
		-e "y_label='nanoseconds'" \
		-e "column_1=1" \
		-e "column_2=4" \
		"$dir/performance.gp"

	# Conjplot
	awk '/BenchmarkConj/{count ++; sub(/BenchmarkConj/,""); printf("%d,%s,%s,%s\n",count,$1,$2,$3)}' "$data" > "$dir/conj.dat"
	gnuplot \
		-e "file_path='$dir/conj.dat'" \
		-e "graphic_file_name='$dir/conj.png'" \
		-e "graphic_title='Conj()'" \
		-e "y_label='nanoseconds'" \
		-e "column_1=1" \
		-e "column_2=4" \
		"$dir/performance.gp"

	# Assoc plot
	awk '/BenchmarkAssoc/{count ++; sub(/BenchmarkAssoc/,""); printf("%d,%s,%s,%s\n",count,$1,$2,$3)}' "$data" > "$dir/assoc.dat"
	gnuplot \
		-e "file_path='$dir/assoc.dat'" \
		-e "graphic_file_name='$dir/assoc.png'" \
		-e "graphic_title='Assoc()'" \
		-e "y_label='nanoseconds'" \
		-e "column_1=1" \
		-e "column_2=4" \
		"$dir/performance.gp"
}

mkdir -p "$dir"

echo "Running benchmarks to generate $data"
go test -bench=. | tee "$data"

printf "benchmark data found in $data, generating plots..." --
plot
echo " done!"
