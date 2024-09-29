package main

import "github.com/charmbracelet/bubbletea"
import "time"
import "fmt"

func write_screen_goncurses(stdscr *goncurses.Window,
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
