#!bash
bootenv() {
    if [ $# -eq 0 ]; then
        if command -v grub-editenv >/dev/null; then
            grub-editenv list
        else
            fw_printenv
        fi
    else
        if command -v grub-editenv >/dev/null; then
            grub-editenv list | grep "^$1"
        else
            fw_printenv "$1"
        fi | sed "s/^${1}=//"
    fi
}
