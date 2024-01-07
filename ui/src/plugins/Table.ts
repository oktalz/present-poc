// zbTable.ts
import MarkdownIt from 'markdown-it';
import StateBlock from 'markdown-it/lib/rules_block/state_block';
import Token from 'markdown-it/lib/token';

export default function zbTable(md: MarkdownIt) {
    //md.inline.ruler.before('text', 'zb_table', function(state: StateInline, silent: boolean) {
    md.block.ruler.before('paragraph', 'zb_table', function(state: StateBlock, startLine: number, endLine: number, silent: boolean) {
        let pos = state.bMarks[startLine] + state.tShift[startLine];
        let max = state.eMarks[startLine];

        // Check if the block starts with '.table'
        if (pos + 6 > max) return false;
        if (state.src.slice(pos, pos + 6) !== '.table') return false;

        let line = state.getLines(startLine, startLine+1, state.blkIndent, false);
        let lineNum = startLine+1
        let table = `<table>`
        let trOpen = false
        let tdContent = ""
        for(;line!='.table.end' && lineNum<endLine; lineNum++){
            line = state.getLines(lineNum, lineNum+1, state.blkIndent, false);
            // console.log(line);
            if (line == '.table.end') {
                if (tdContent != "") {
                    const content = md.render(tdContent);
                    table += `<td>`+content+`</td>`;
                }
                table += `</tr>`
                table += `</table>`
                break
            }
            if (line == '.tr') {
                if (trOpen) {
                    if (tdContent != "") {
                        const content = md.render(tdContent);
                        table += `<td>`+content+`</td>`;
                        tdContent = ""
                    }
                    table += `</tr>`
                }
                trOpen = true;
                table += `<tr>`
            } else if (line.startsWith(".td")) {
                if (tdContent != "") {
                    const content = md.renderInline(tdContent);
                    table += `<td>`+content+`</td>`;
                    tdContent = ""
                }
                tdContent = line.slice(3)
            } else {
                tdContent += "\n"+line
            }
        }
        if (table == `<table>`) {
            return false
        }
        if (!silent) {
            const token: Token = state.push('zb_table', '', 0);
            token.content = table;
            //token.markup = `.table${match}.table.end`;

        }
        state.line = lineNum + 1;
        return true;

        /*rows.forEach(element => {
            table += `<tr>`;
            const columns = element.split('\n.td');
            columns.shift(); // Skip the first element
            columns.forEach(element => {
                const content = md.renderInline(element);
                table += `<td>`+content+`</td>`;
            });
            table += `</tr>`;
        });
        table += `</table>`;
        */

        if (!silent) {
            const token: Token = state.push('zb_table', '', 0);
            token.content = table;
            //token.markup = `.table${match}.table.end`;
        }

        //state.pos = end + 1 + 9; // 8 = table.end
        return true;
    });

    md.renderer.rules['zb_table'] = function(tokens: Token[], idx: number) {
        return tokens[idx].content;
    };
};
