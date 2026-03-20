package drift

// zshSnippet is emitted by `drift shell-init zsh`.
//
// Mechanism: zsh's TMOUT/TRAPALRM system.
// TMOUT sets an idle alarm; TRAPALRM fires instead of exiting the shell.
// After the trap returns, zsh resets the timer for another TMOUT seconds.
const zshSnippet = `# drift – terminal screensaver (zsh integration)
# https://github.com/phlx0/drift

_drift_activate() {
  # Guard: only fire when the shell is idle at a prompt.
  # ZLE_STATE is set when zsh is waiting for line editor input.
  [[ -z "${DRIFT_ENABLED}" ]] && return
  [[ -n "${DRIFT_RUNNING}" ]] && return

  export DRIFT_RUNNING=1
  command -v drift >/dev/null 2>&1 && drift
  unset DRIFT_RUNNING
}

# Seconds of inactivity before drift activates.
# Override by setting DRIFT_TIMEOUT in your shell config before sourcing this.
TMOUT="${DRIFT_TIMEOUT:-120}"

# TRAPALRM is called by zsh when the TMOUT alarm fires.
# Defining it prevents zsh from exiting on timeout.
TRAPALRM() {
  _drift_activate
}

export DRIFT_ENABLED=1
`

// bashSnippet is emitted by `drift shell-init bash`.
//
// Mechanism: PROMPT_COMMAND launches a background timer process on each
// new prompt.  The timer calls drift when it expires.  Running a command
// cancels the current timer via a PROMPT_COMMAND at the start of the cycle.
const bashSnippet = `# drift – terminal screensaver (bash integration)
# https://github.com/phlx0/drift

_drift_cancel() {
  if [[ -n "${_DRIFT_PID}" ]]; then
    kill "${_DRIFT_PID}" 2>/dev/null
    unset _DRIFT_PID
  fi
}

_drift_schedule() {
  local timeout="${DRIFT_TIMEOUT:-120}"
  set +m
  (
    sleep "${timeout}"
    [[ "$(ps -o stat= -p $$)" == *"+"* ]] || exit 0
    command -v drift >/dev/null 2>&1 && drift
  ) &
  _DRIFT_PID=$!
  set -m
  disown "${_DRIFT_PID}" 2>/dev/null
}

_drift_prompt() {
  _drift_cancel
  _drift_schedule
}

# Prepend our hook so it runs before any existing PROMPT_COMMAND.
if [[ -z "${PROMPT_COMMAND}" ]]; then
  PROMPT_COMMAND="_drift_prompt"
else
  PROMPT_COMMAND="_drift_prompt; ${PROMPT_COMMAND}"
fi
`

// fishSnippet is emitted by `drift shell-init fish`.
//
// Mechanism: fish event functions.
// fish_prompt fires when a new prompt is drawn (= user is idle at the prompt).
// fish_preexec fires before any command runs (= user is active).
const fishSnippet = `# drift – terminal screensaver (fish integration)
# https://github.com/phlx0/drift

function _drift_cancel
  if set -q _drift_timer_pid
    kill $_drift_timer_pid 2>/dev/null
    set -e _drift_timer_pid
  end
end

function _drift_schedule
  set -l timeout (set -q DRIFT_TIMEOUT; and echo $DRIFT_TIMEOUT; or echo 120)
  fish -c "sleep $timeout; and command -s drift >/dev/null 2>&1; and drift" &
  set -g _drift_timer_pid $last_pid
end

function _drift_on_prompt --on-event fish_prompt
  _drift_cancel
  _drift_schedule
end

function _drift_on_preexec --on-event fish_preexec
  _drift_cancel
end
`
