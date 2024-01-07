// zbStyle.ts
import MarkdownIt from 'markdown-it';
import StateBlock from 'markdown-it/lib/rules_block/state_block';
import Token from 'markdown-it/lib/token';

export default function zbStyle(md: MarkdownIt) {
    md.block.ruler.before('paragraph', 'zb_style', function(state: StateBlock, startLine: number, endLine: number, silent: boolean) {
        let pos = state.bMarks[startLine] + state.tShift[startLine];
        let max = state.eMarks[startLine];

        // Check if the block starts with '.style'
        if (pos + 6 > max) return false;
        if (state.src.slice(pos, pos + 6) !== '.style') return false;

        let firstLine = state.getLines(startLine, startLine+1, state.blkIndent, false);
        // check if we have just one line
        let parts = firstLine.split(`"`);
        parts = parts.filter(part => part !== "");
        // console.log(parts);
        if (parts.length > 2) {
            // OK we are in single line mode
            if (!silent) {
                const token: Token = state.push('zb_style', 'div', 0);
                token.map = [startLine, startLine+1];
                token.attrSet('style', parts[1].trim());
                parts.splice(0, 2);
                token.content = md.render(parts.join(` `));  // Process the rest as regular markdown
            }
            state.line = startLine + 1;
            return true;
        }


        // Check if the block ends with '.style.end'
        let nextLine = startLine + 1;
        while (nextLine < endLine) {
            if (state.sCount[nextLine] < state.blkIndent) break;
            pos = state.bMarks[nextLine] + state.tShift[nextLine];
            max = state.eMarks[nextLine];
            if (pos + 10 <= max && state.src.slice(pos, pos + 10) === '.style.end') {
                // Found the end marker, so we can process the block
                if (!silent) {
                    const token: Token = state.push('zb_style', 'div', 0);
                    token.map = [startLine, nextLine];
                    firstLine = firstLine.replace(".style ", "").trim();
                    if (firstLine.startsWith('"') && firstLine.endsWith('"')) {
                        firstLine = firstLine.slice(1, -1).trim();
                    }

                    token.attrSet('style', firstLine);
                    //const lines = state.getLines(startLine + 1, nextLine, state.blkIndent, false).split('\n');
                    //token.attrSet('style', lines.shift().trim());  // Set the style attribute
                    token.content = md.render(state.getLines(startLine + 1, nextLine, state.blkIndent, false));  // Process the rest as regular markdown
                }
                state.line = nextLine + 1;
                return true;
            }
            nextLine++;
        }

        // If we're here, the block didn't end with '.style.end', so we don't process it
        return false;
    });

    md.renderer.rules['zb_style'] = function(tokens: Token[], idx: number) {
        const style = tokens[idx].attrGet('style');
        return `<div style="${style}">${tokens[idx].content}</div>`;
    };
};
