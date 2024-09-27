package main

import "fmt"
import "log"
import "bufio"
import "os"
import "strings"
import "time"
import "github.com/gbin/goncurses"
import "path/filepath"

func ReadNoOfSets(stdscr *goncurses.Window) int {
	var err error = fmt.Errorf("Init no error")
	var nos int = 0

	for err != nil {
		stdscr.ColorOn(goncurses.C_WHITE)
		print_to_box(stdscr, "      How many sets? (Enter 'q' to quit)", 3)
		print_to_box(stdscr, "            ", 4)
		refresh_box(stdscr)

		input := stdscr.GetChar()
		stdscr.ColorOff(goncurses.C_WHITE)

		if input == 'q' {
			nos = -1
			break
		}
		// TODO: For some reason, pressing delete will result in an input value of 51,
		// which makes nos = 3. del key is suppposed to be 127 though. No idea why that happens.
		if int(input) <= int('9') && int(input) >= int('0') {
			nos = int(input) - int('0')
			break
		}
	}
	return nos
}

func ReadExercises() []string {
	readFile, err := os.Open("/home/andreas/projects/workout_timer_go/workout_list")

	if err != nil {
		log.Fatal(err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}
	readFile.Close()

	return fileLines
}

func sum_times(array_to_sum_up []time.Duration) time.Duration {
	var total_time time.Duration

	for _, time_frame := range array_to_sum_up {
		total_time += time_frame
	}
	return total_time
}

func write_diagram(i_count int, segment string, nos int) string {
	i_str := ""
	retstr := ""

	if segment == "workout" {
		i_str = "I "
	} else if segment == "break" {
		i_str = " I"
	}

	for i := 0; i < i_count; i++ {
		if i != 0 && i%nos == 0 {
			retstr = fmt.Sprintf("%s%s", retstr, " | ")
		}
		retstr = fmt.Sprintf("%s%s", retstr, i_str)
	}

	return retstr
}

func duration_to_string(duration time.Duration) string {
	return fmt.Sprintf("%02d:%02d:%02d", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)
}

func print_to_box(box *goncurses.Window, line string, cur_line int) int {
	box.MovePrint(cur_line, 2, line)
	return cur_line + 1

}

func write_screen(stdscr *goncurses.Window,
	t_time_elapsed time.Duration,
	w_time_elapsed time.Duration,
	b_time_elapsed time.Duration,
	c_time_elapsed time.Duration,
	c_segment string,
	w_count int,
	b_count int,
	len_workouts int,
	workouts []string,
	nos int,
	quit_next bool) []string {

	cur_line := 1

	lines := []string{}
	lines = append(lines, fmt.Sprintf("Total time elapsed      %s", duration_to_string(t_time_elapsed)))
	lines = append(lines, "---------------------------------------")
	lines = append(lines, fmt.Sprintf("Workout time elapsed:   %s   %s", duration_to_string(w_time_elapsed), write_diagram(w_count, "workout", len_workouts)))
	lines = append(lines, fmt.Sprintf("Break time elapsed      %s   %s", duration_to_string(b_time_elapsed), write_diagram(b_count, "break", len_workouts)))
	lines = append(lines, "---------------------------------------")
	lines = append(lines, fmt.Sprintf("%d %% done", (w_count*100/(len_workouts*nos))))

	cur_time_line := fmt.Sprintf("Current %s: %s\n", c_segment, duration_to_string(c_time_elapsed))

	var color_pair int16 = goncurses.C_WHITE
	if c_segment == "break" {
		if c_time_elapsed.Seconds() >= 3*60 {
			color_pair = goncurses.C_GREEN
		} else if c_time_elapsed.Seconds() >= 1*60 {
			color_pair = goncurses.C_YELLOW
		} else {
			color_pair = goncurses.C_RED
		}
	} else {
		color_pair = goncurses.C_WHITE
	}
	stdscr.Clear()

	stdscr.ColorOn(goncurses.C_WHITE)
	for _, line := range lines {
		cur_line = print_to_box(stdscr, line, cur_line)
	}
	stdscr.ColorOff(goncurses.C_WHITE)

	if !quit_next {
		stdscr.ColorOn(color_pair)
		cur_line = print_to_box(stdscr, cur_time_line, cur_line)
		stdscr.ColorOff(color_pair)

		stdscr.ColorOn(goncurses.C_WHITE)
		var cur_segment_line string

		cur_workout := workouts[w_count%len_workouts]
		if c_segment == "break" {
			cur_segment_line = fmt.Sprintf("Get Ready for %s", cur_workout)
		} else {
			cur_segment_line = fmt.Sprintf("Doing %s", cur_workout)
		}
		cur_line = print_to_box(stdscr, cur_segment_line, cur_line)
		stdscr.ColorOff(goncurses.C_WHITE)
	}

	refresh_box(stdscr)

	if quit_next {
		cur_line = print_to_box(stdscr, "Press any key to exit", cur_line)
		stdscr.GetChar()
	}

	return lines
}

func find_workout_len(workouts []string) int {
	nonempty_lines := 0
	for _, workout := range workouts {
		if len(strings.TrimSpace(workout)) > 0 {
			nonempty_lines++
		}
	}
	fmt.Printf("%d", nonempty_lines)
	return nonempty_lines
}

func RunWorkoutLoop(stdscr *goncurses.Window, nos int, workouts []string, c chan goncurses.Key) ([]string, int, time.Duration, time.Duration, time.Duration) {
	workout_start := time.Now()
	var w_time_elapsed_arr []time.Duration
	var b_time_elapsed_arr []time.Duration
	len_workouts := find_workout_len(workouts)
	segment := "workout"
	last_kh := time.Now()

	quit_next := false
	last_screen := []string{}

	for true {
		time.Sleep(time.Millisecond * 100)
		c_time_elapsed := time.Now().Sub(last_kh)

		if segment == "workout" {
			last_screen = write_screen(stdscr,
				time.Now().Sub(workout_start),
				sum_times(w_time_elapsed_arr)+c_time_elapsed,
				sum_times(b_time_elapsed_arr),
				c_time_elapsed,
				segment,
				len(w_time_elapsed_arr),
				len(b_time_elapsed_arr),
				len_workouts,
				workouts,
				nos,
				quit_next)
		} else if segment == "break" {
			last_screen = write_screen(stdscr,
				time.Now().Sub(workout_start),
				sum_times(w_time_elapsed_arr),
				sum_times(b_time_elapsed_arr)+c_time_elapsed,
				c_time_elapsed,
				segment,
				len(w_time_elapsed_arr),
				len(b_time_elapsed_arr),
				len_workouts,
				workouts,
				nos,
				quit_next)
		}

		if quit_next {
			break
		}

		select {
		case v := <-c:
			if v == 'q' {
				quit_next = true
			} else if v == 'n' {
				if segment == "workout" {
					segment = "break"
					w_time_elapsed_arr = append(w_time_elapsed_arr, time.Now().Sub(last_kh))
				} else if segment == "break" {
					segment = "workout"
					b_time_elapsed_arr = append(b_time_elapsed_arr, time.Now().Sub(last_kh))
				}
				last_kh = time.Now()

				if len(w_time_elapsed_arr) >= len_workouts*nos {
					quit_next = true
				}
			}
		default:
		}
	}
	return last_screen, len(w_time_elapsed_arr), sum_times(w_time_elapsed_arr), sum_times(b_time_elapsed_arr), sum_times(w_time_elapsed_arr) + sum_times(b_time_elapsed_arr)
}

func check_err(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func run_input_loop(stdscr *goncurses.Window, c chan goncurses.Key) {
	// Read input
	for {
		char := stdscr.GetChar()
		stdscr.MovePrint(0, 0, char)

		if char == 'b' || char == 'n' || char == 's' || char == 'q' {
			c <- char
		}

		if char == 'q' {
			break
		}
	}
	close(c)
}

func refresh_box(win *goncurses.Window) {
	win.Box(0, 0)
	win.Refresh()
}

func save_to_file(lines []string, no_exercises int, no_done_workouts int, w_time time.Duration, b_time time.Duration, t_time time.Duration) {

	dy := time.Now().Day()
	mo := time.Now().Month()
	yr := time.Now().Year()

	hr := time.Now().Hour()
	mn := time.Now().Minute()

	dir_name := "/home/andreas/projects/workout_timer_go/done_workouts"

	_, err := os.Stat(dir_name)

	if err != nil {
		err = os.Mkdir(dir_name, 0775)
		if err != nil {
			log.Fatal(err)
		}
	}

	filename := filepath.Join(dir_name, fmt.Sprintf("%02d_%02d", yr, mo))

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	write_to_file(f, "---------------------------\n")
	write_to_file(f, fmt.Sprintf("Workout finished on: %02d.%02d.%04d, %02d:%02d\n", dy, mo, yr, hr, mn))
	write_to_file(f, fmt.Sprintf("Workouts with number of sets each:\n"))
	write_to_file(f, "---------------------------\n")

	for i, line := range lines {
		wo_count := int(no_done_workouts / no_exercises)
		if i < no_done_workouts%no_exercises {
			wo_count++
		}
		if wo_count > 0 {
			write_to_file(f, fmt.Sprintf("%s: %d\n", line, wo_count))
		}
	}

	write_to_file(f, "---------------------------\n")
	write_to_file(f, fmt.Sprintf("%d workouts total\n", no_done_workouts))
	write_to_file(f, fmt.Sprintf("%s workout time total\n", duration_to_string(w_time)))
	write_to_file(f, fmt.Sprintf("%s break time total \n", duration_to_string(b_time)))
	write_to_file(f, fmt.Sprintf("%s time total\n", duration_to_string(t_time)))
	write_to_file(f, "---------------------------\n")
}

func write_to_file(f *os.File, line string) {
	if _, err := f.WriteString(line); err != nil {
		log.Fatal(err)
	}
}

func print_to_screen(lines []string) {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		fmt.Println(line)
	}
}

func intro_screen(win *goncurses.Window) {
	curline := 1
	win.Clear()
	curline = print_to_box(win, "", curline)
	curline = print_to_box(win, "", curline)
	curline = print_to_box(win, "                 Welcome to Workout", curline)
	curline = print_to_box(win, "            Press any Key to continue", curline)
	refresh_box(win)

	win.GetChar()
}

func workout() []string {
	exercises := ReadExercises()

	_, err := goncurses.Init()
	check_err(err)
	defer goncurses.End() // Clean up on exit

	if err := goncurses.StartColor(); err != nil {
		log.Fatal(err)
	}
	goncurses.UseDefaultColors()

	goncurses.InitPair(goncurses.C_GREEN,
		goncurses.C_GREEN, -1)
	goncurses.InitPair(goncurses.C_YELLOW,
		goncurses.C_YELLOW, -1)
	goncurses.InitPair(goncurses.C_RED,
		goncurses.C_RED, -1)
	goncurses.InitPair(goncurses.C_WHITE,
		255, -1)

	win, err := goncurses.NewWindow(10, 80, 5, 5)
	check_err(err)
	refresh_box(win)

	nos := ReadNoOfSets(win)

	if nos == -1 {
		return []string{}
	}

	intro_screen(win)

	c := make(chan goncurses.Key, 1)
	go run_input_loop(win, c)

	last_screen, no_done_workouts, w_time, b_time, t_time := RunWorkoutLoop(win, nos, exercises, c)
	if no_done_workouts > 0 {
		save_to_file(exercises, find_workout_len(exercises), no_done_workouts, w_time, b_time, t_time)
	}

	return last_screen
}

func main() {
	workout()
}
