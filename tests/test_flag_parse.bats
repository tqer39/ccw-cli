#!/usr/bin/env bats

CCW="${BATS_TEST_DIRNAME}/../bin/ccw"

@test "-h prints usage" {
  run "$CCW" -h
  [ "$status" -eq 0 ]
  [[ "$output" == *"Usage: ccw"* ]]
}

@test "--help prints usage" {
  run "$CCW" --help
  [ "$status" -eq 0 ]
  [[ "$output" == *"Usage: ccw"* ]]
}

@test "-v prints version" {
  run "$CCW" -v
  [ "$status" -eq 0 ]
  [[ "$output" == "ccw "* ]]
}

@test "unknown option errors out" {
  run "$CCW" --nope
  [ "$status" -eq 1 ]
  [[ "$output" == *"unknown option: --nope"* ]]
}

@test "unknown argument errors out" {
  run "$CCW" foobar
  [ "$status" -eq 1 ]
  [[ "$output" == *"unknown argument: foobar"* ]]
}
