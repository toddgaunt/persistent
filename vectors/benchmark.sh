#!/usr/bin/env bash

dir=benchmark
data=$dir/benchmark.dat

plot() {
	# NthTrie plot
	awk '/BenchmarkNthTrie/{count ++; sub(/BenchmarkNthTrie/,""); printf("%d,%s,%s,%s\n",count,$1,$2,$3)}' "$data" > "$dir/nth_trie.dat"
	gnuplot \
		-e "file_path='$dir/nth_trie.dat'" \
		-e "graphic_file_name='$dir/nth_trie.png'" \
		-e "graphic_title='Nth in Trie'" \
		-e "y_label='nanoseconds'" \
		-e "y_range_min='0'" \
		-e "y_range_max='50'" \
		-e "column_1=1" \
		-e "column_2=4" \
		"$dir/performance.gp"


	# NthTail plot
	awk '/BenchmarkNthTail/{count ++; sub(/BenchmarkNthTail/,""); printf("%d,%s,%s,%s\n",count,$1,$2,$3)}' "$data" > "$dir/nth_tail.dat"
	gnuplot \
		-e "file_path='$dir/nth_tail.dat'" \
		-e "graphic_file_name='$dir/nth_tail.png'" \
		-e "graphic_title='Nth in Tail (fast)'" \
		-e "y_label='nanoseconds'" \
		-e "y_range_min='0'" \
		-e "y_range_max='20'" \
		-e "column_1=1" \
		-e "column_2=4" \
		"$dir/performance.gp"

	# AssocTrie plot
	awk '/BenchmarkAssocTrie/{count ++; sub(/BenchmarkAssocTrie/,""); printf("%d,%s,%s,%s\n",count,$1,$2,$3)}' "$data" > "$dir/assoc_trie.dat"
	gnuplot \
		-e "file_path='$dir/assoc_trie.dat'" \
		-e "graphic_file_name='$dir/assoc_trie.png'" \
		-e "graphic_title='Assoc in Trie'" \
		-e "y_label='nanoseconds'" \
		-e "y_range_min='0'" \
		-e "y_range_max='8000'" \
		-e "column_1=1" \
		-e "column_2=4" \
		"$dir/performance.gp"


	# AssocTail plot
	awk '/BenchmarkAssocTail/{count ++; sub(/BenchmarkAssocTail/,""); printf("%d,%s,%s,%s\n",count,$1,$2,$3)}' "$data" > "$dir/assoc_tail.dat"
	gnuplot \
		-e "file_path='$dir/assoc_tail.dat'" \
		-e "graphic_file_name='$dir/assoc_tail.png'" \
		-e "graphic_title='Assoc in Tail (fast)'" \
		-e "y_label='nanoseconds'" \
		-e "y_range_min='0'" \
		-e "y_range_max='200'" \
		-e "column_1=1" \
		-e "column_2=4" \
		"$dir/performance.gp"

	# ConjTrie plot
	awk '/BenchmarkConjTrie/{count ++; sub(/BenchmarkConjTrie/,""); printf("%d,%s,%s,%s\n",count,$1,$2,$3)}' "$data" > "$dir/conj_trie.dat"
	gnuplot \
		-e "file_path='$dir/conj_trie.dat'" \
		-e "graphic_file_name='$dir/conj_trie.png'" \
		-e "graphic_title='Conj in Trie'" \
		-e "y_label='nanoseconds'" \
		-e "y_range_min='0'" \
		-e "y_range_max='8000'" \
		-e "column_1=1" \
		-e "column_2=4" \
		"$dir/performance.gp"

	# ConjTail plot
	awk '/BenchmarkConjTail/{count ++; sub(/BenchmarkConjTail/,""); printf("%d,%s,%s,%s\n",count,$1,$2,$3)}' "$data" > "$dir/conj_tail.dat"
	gnuplot \
		-e "file_path='$dir/conj_tail.dat'" \
		-e "graphic_file_name='$dir/conj_tail.png'" \
		-e "graphic_title='Conj in Tail (fast)'" \
		-e "y_label='nanoseconds'" \
		-e "y_range_min='0'" \
		-e "y_range_max='500'" \
		-e "column_1=1" \
		-e "column_2=4" \
		"$dir/performance.gp"
}

mkdir -p "$dir"

if ! [[ -f "$data" ]]; then
	echo "Running benchmarks to generate $data"
	go test -bench=. | tee "$data"
fi

printf "benchmark data found in $data, generating plots..." --
plot
echo " done!"
