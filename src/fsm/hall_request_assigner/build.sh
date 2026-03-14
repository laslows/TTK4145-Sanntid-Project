#!/usr/bin/env bash
set -e

gdc main.d config.d elevator_algorithm.d elevator_state.d optimal_hall_requests.d d-json/jsonx.d -Wall -g -o hall_request_assigner