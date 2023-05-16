# unscramble.go

Take some letters and find words that can be made with those letters

## What?
### unscramble.go
This is my first attempt at Go.  I have no idea the conventions or best practices, so feel free to comment.\ 
I decided to try and create something similar to my boggle solver, particularly because I wanted an easy way to unscramble some words\ 
...And yes, I know there are utilities, but 1.  Nothing more fun than doing to yourself, 2. I can feed it any words that I want

Example:
```bash
% go run /Users/syoung/git2/unscramble/unscramble.go  --letters exmaple --min 5 --max 5
```
``` ignorelang
 Starting
 Finding words of 5 to 5 length
 Found 4 words
 expel
 maple
 ample
 pelma
 Done
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
- Fixed annoying blank space at begining of line