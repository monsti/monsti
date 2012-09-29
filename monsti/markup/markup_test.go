package markup

import (
	"testing"
)

func TestMarkupToHTML(t *testing.T) {
	for i := 0; i < len(markupTests); i = i + 2 {
		out := MarkupToHTML(markupTests[i])
		if out != markupTests[i+1] {
		msg := `MarkupToHTML(...) returned wrong response:
========= IN ===========
%v
========= OUT ==========
%v
========= WANT =========
%v
========================`
		t.Errorf(msg, markupTests[i], out, markupTests[i+1])
		}
	}
}

var markupTests []string = []string{`# Header1
Header1
=======
## Header 2
Header 2
--------
### Header 3
#### Header 4
##### Header 5
###### Header 6
`,
`<h1>Header1</h1>

<h1>Header1</h1>

<h2>Header 2</h2>

<h2>Header 2</h2>

<h3>Header 3</h3>

<h4>Header 4</h4>

<h5>Header 5</h5>

<h6>Header 6</h6>
`,`Paragraph one. Another sentence.

Paragraph two. Yet another
sentence.
`,
`<p>Paragraph one. Another sentence.</p>

<p>Paragraph two. Yet another
sentence.</p>
`,`* Item one.
    + Subitem.
* Item two.

Separate them!

1. One
2. Two
`,
`<ul>
<li>Item one.

<ul>
<li>Subitem.</li>
</ul></li>
<li>Item two.</li>
</ul>

<p>Separate them!</p>

<ol>
<li>One</li>
<li>Two</li>
</ol>
`,`This paragraph has a  
line break.
`,
`<p>This paragraph has a<br />
line break.</p>
`,`Testing *italic* or _italic_.
Testing **bold** or __bold__.
But t_hi_s should not work.
Go ~~strike this~~.
`,
`<p>Testing <em>italic</em> or <em>italic</em>.
Testing <strong>bold</strong> or <strong>bold</strong>.
But t_hi_s should not work.
Go <del>strike this</del>.</p>
`,`"Smarty pants!" 'More smarty pants!' - -- ---
`,
`<p>&ldquo;Smarty pants!&rdquo; &lsquo;More smarty pants!&rsquo; - &ndash; &mdash;</p>
`,`Here is some ` + "`code`" + `.

    10 REM Code!
	20 NOP
	30 GOTO 20

` + "``` go" + `
func moreCode() bool {
	return true
}
` + "```",
`<p>Here is some <code>code</code>.</p>

<pre><code>10 REM Code!
20 NOP
30 GOTO 20
</code></pre>

<pre><code class="go">func moreCode() bool {
    return true
}
</code></pre>
`,`
Column One          | Column Two
--------------------|------------------------
Foo                 | 1
Bar                 | 50
`,
`<table>
<thead>
<tr>
<td>Column One</td>
<td>Column Two</td>
</tr>
</thead>

<tbody>
<tr>
<td>Foo</td>
<td>1</td>
</tr>

<tr>
<td>Bar</td>
<td>50</td>
</tr>
</tbody>
</table>
`,"A fraction: 2/3!", "<p>A fraction: <sup>2</sup>&frasl;<sub>3</sub>!</p>\n",
`> This is a
blockquote paragraph.`,`<blockquote>
<p>This is a
blockquote paragraph.</p>
</blockquote>
`,`Go visit this [nice page](http://example.com) or [this page][linkref].

http://foo.example.com

[linkref]: http://example.com "Nice page!"
`,
`<p>Go visit this <a href="http://example.com">nice page</a> or <a href="http://example.com" title="Nice page!">this page</a>.</p>

<p><a href="http://foo.example.com">http://foo.example.com</a></p>
`,`
--------------
* * * * * * * * *
***
- - -
`,
`<hr />

<hr />

<hr />

<hr />
`}
