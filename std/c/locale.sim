pub type lconv struct{}

pub let LC_CTYPE: int = 0
pub let LC_NUMERIC: int = 1
pub let LC_TIME: int = 2
pub let LC_COLLATE: int = 3
pub let LC_MONETARY: int = 4
pub let LC_MESSAGES: int = 5
pub let LC_ALL: int = 6
pub let LC_PAPER: int = 7
pub let LC_NAME: int = 8
pub let LC_ADDRESS: int = 9
pub let LC_TELEPHONE: int = 10
pub let LC_MEASUREMENT: int = 11
pub let LC_IDENTIFICATION: int = 12

@extern(setlocale)
pub func setlocale(category: int, locale: *char)*char

@extern(localeconv)
pub func localeconv()*lconv