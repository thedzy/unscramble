# unscramble.go

Take some letters and find words that can be made with those letters

## What?
### unscramble.go
This is my first attempt at Go.  I have no idea the conventions or best practices, so feel free to comment.\ 
I decided to try and create something similar to my boggle solver, particularly because I wanted an easy way to unscramble some words\ 
...And yes, I know there are utilities, but 1.  Nothing more fun than doing to yourself, 2. I can feed it any words that I want

Example:
```bash
% go run ./unscramble.go  --letters exmaple --min 5 --max 5
```
``` ignorelang
Starting
Loading 279,496 words into dictionary
Finding words of 5 to 5 length
Found 4 words
Displaying 4 words
----
expel
maple
ample
pelma
----
Done
```

Example:
```bash
% go run ./unscramble.go  -l ebcdfghjklmnpqrstvwxz --min 7 -j 2>/dev/null | jq .
```
``` json
[
  "klephts",
  "lengths",
  "phlegms",
  "pschent",
  "schmelz"
]
```
Example:
```bash
echo "aeiouwx" | go run ./unscramble.go --sort l --limit 10 2>/dev/null
```
``` ignorelang
ae
xi
ou
aw
we
ax
wo
ox
xu
ai
```
Example:
```bash
echo aabcdfghjklmnpqrstvwxyz | ./unscramble --log-level 30 -j --filter "^
a.*[ty]$"
```
``` json
["abaft","ably","abray","aby","achy","act","adapt","adry","aft","agast","aghast","agly","alant","alary","alay","alky","alt","alway","ambary","ambry","amply","analyst","anarchy","angary","angry","angst","angsty","ant","antsy","any","apart","apathy","apay","apt","aptly","arblast","archly","archway","arhat","army","arsy","art","artsy","arty","ary","ashplant","ashtray","ashy","askant","aslant","asphalt","astray","asway","at","ataxy","avant","avast","away","awfy","awmry","awny","awry","ay"]
```

## Why?
Kinda covered this above, but mostly to try Go

## Improvements?
Speed wise I think it's about as fast as it gets.  Going to look at multithreading, but I think Go already does that

## State?
No known bugs.  Works.  

## New
### 1.0
- Takes some letter and rearrange them to new words
- Human-readable or json output
### 1.1
- Bug fixes, not assuming there is a terminating character on the lines and adding one
- Adding option to specify existing terminating characters on lines
- Spelling fixes and code cleanup
### 1.2
- Better fix for the terminating character, splitting on any and all control characters
- Outputting to stderr, except words, so that they can be redirected
- Code cleanup after finding some suggestions
- Grammar!
### 1.2.1
- Fixed annoying blank space at beginning of line
### 1.3
- Added stdin
- Added a limit option
### 1.4
- Fix: Opening files as read only
- Print amount of words we are comparing against
- Improved debugging message
- Prevent zero length string from matching
### 1.5
- Added option to filter results with a regex
- Fix: Empty results in json does not output null
- Allow all by control characters in letters
- Code improvements
