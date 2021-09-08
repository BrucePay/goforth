" Vim syntax file
" Language:           GoForth (a contatenative language written in Go; heavily influenced by Joy
" Maintainer:         Bruce Payette
" Version:            0.1
" Project Repository: https://github.com/brucepay
" Vim Script Page:    http://www.vim.org/scripts/
"
" The following settings are available for tuning syntax highlighting:
"    let braid_nofold_blocks = 1
"    let braid_nofold_region = 1

" Compatible VIM syntax file start
if version < 600
	syntax clear
elseif exists("b:current_syntax")
	finish
endif

" Operators contain dashes
setlocal iskeyword+=-

" Braid does't care about case
" syn case ignore

" Sync-ing method
syn sync minlines=100

" Certain tokens can't appear at the top level of the document
syn cluster braidNotTop contains=@braidComment,braidCDocParam,braidFunctionDeclaration

" Comments and special comment words
syn keyword braidCommentTodo  TODO FIXME XXX TBD HACK NOTE BUGBUG BUGBUGBUG contained
syn match braidComment        /(;.*;)\|#.*/ contains=goforthCommentTodo,goforthCommentDoc,@Spell

" Language keywords and elements
syn keyword goforthKeyword     DEFINE == if ifte while map each reduce repeat case primrec linrec binrec dip

syn keyword goforthConstant    true false null nil _  IsLinux IsMacOS IsWindows IsCoreCLR IsUnix tid 


" Variable references
syn match goforthVariable      /\w\+/ 

" Type literals
syn match goforthType /\^[a-z_][a-z0-9_.,\[\]]*/

" goforth Operators
syn keyword goforthOperator is? as number? list? nil? null? lambda? atom? symbol? string? bound? dict?
syn keyword goforthOperator keyword? pair? quote? zero? band bor not and or
syn match goforthOperator /[a-z_][._a-z0-9]*\/[a-z_][a-z0-9_]*/
syn match goforthOperator /\./
syn match goforthOperator /=/
syn match goforthOperator /+/
syn match goforthOperator /\*/
syn match goforthOperator /\*\*/
syn match goforthOperator /\//
syn match goforthOperator /|/
syn match goforthOperator /%/
syn match goforthOperator /,/
syn match goforthOperator /\./
syn match goforthOperator /\.\./
syn match goforthOperator /</
syn match goforthOperator /<=/
syn match goforthOperator />/
syn match goforthOperator />=/
syn match goforthOperator /==/
syn match goforthOperator /!=/
syn match goforthOperator /->/
syn match goforthOperator /\.[a-z_][._a-z0-9]*/
syn match goforthOperator /\.[a-z_][._a-z0-9]*\/[a-z_][a-z0-9_]*/
syn match goforthOperator /?\[/
syn match goforthOperator /\~/
syn match goforthOperator /\[/
syn match goforthOperator /\]/
syn match goforthOperator /(/
syn match goforthOperator /)/


" Regular expression literals
syn region goforthString start=/r\// skip=/\\\// end=/\//

" Strings
syn region goforthString start=/"/ skip=/\\"/ end=/"/ contains=@Spell


" Interpolation in strings
syn region goforthInterpolation matchgroup=goforthInterpolationDelimiter start="${" end="}" contained contains=ALLBUT,@goforthNotTop
syn region goforthNestedParentheses start="(" skip="\\\\\|\\)" matchgroup=goforthInterpolation end=")" transparent contained
syn cluster goforthStringSpecial contains=goforthEscape,goforthInterpolation,goforthVariable,goforthBoolean,goforthConstant,goforthBuiltIn,@Spell

" Numbers
syn match   goforthNumber		"\(\<\|-\)\@<=\(0[xX]\x\+\|\d\+\)\([KMGTP][B]\)\=\(\>\|-\)\@="
syn match   goforthNumber		"\(\(\<\|-\)\@<=\d\+\.\d*\|\.\d\+\)\([eE][-+]\=\d\+\)\=[dD]\="
syn match   goforthNumber		"\<\d\+[eE][-+]\=\d\+[dD]\=\>"
syn match   goforthNumber		"\<\d\+\([eE][-+]\=\d\+\)\=[dD]\>"
syn match   goforthNumber      "\<\d\+i\>" " bigint constants

" Constants
syn match goforthBoolean        "\%(true\|false\)\>"
syn match goforthConstant       /\nil\>/

" Folding blocks
if !exists('g:goforth_nofold_blocks')
	syn region goforthBlock start=/{/ end=/}/ transparent fold
endif

if !exists('g:goforth_nofold_region')
	syn region goforthRegion start=/#region/ end=/#endregion/ transparent fold keepend extend
endif

" Setup default color highlighting
if version >= 508 || !exists("did_goforth_syn_inits")

    if version < 508
		let did_goforth_syn_inits = 1
		command -nargs=+ HiLink hi link <args>
	else
		command -nargs=+ HiLink hi def link <args>
	endif

	HiLink goforthNumber Number
	HiLink goforthBlock Block
	HiLink goforthRegion Region
	HiLink goforthException Exception
	HiLink goforthConstant Constant
	HiLink goforthString String
	HiLink goforthEscape SpecialChar
	HiLink goforthInterpolationDelimiter Delimiter
	HiLink goforthConditional Conditional
	HiLink goforthFunctionDeclaration Function
	HiLink goforthFunctionInvocation Function
	HiLink goforthVariable Identifier
	HiLink goforthBoolean Boolean
	HiLink goforthConstant Constant
	HiLink goforthBuiltIn StorageClass
	HiLink goforthType Type
	HiLink goforthComment Comment
	HiLink goforthCommentTodo Todo
	HiLink goforthCommentDoc Tag
	HiLink goforthCDocParam Todo
	HiLink goforthOperator Operator
	HiLink goforthRepeat Repeat
	HiLink goforthRepeatAndCmdlet Repeat
	HiLink goforthKeyword Keyword
endif

let b:current_syntax = "goforth"
