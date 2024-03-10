// zbTable.ts
import MarkdownIt from 'markdown-it';
import StateBlock from 'markdown-it/lib/rules_block/state_block';
import Token from 'markdown-it/lib/token';

export default function zbTable(md: MarkdownIt) {
    md.block.ruler.before('paragraph', 'zb_tab', function(state: StateBlock, startLine: number, endLine: number, silent: boolean) {
        let pos = state.bMarks[startLine] + state.tShift[startLine];
        let max = state.eMarks[startLine];

        if (pos + 5 > max) return false;
        if (state.src.slice(pos, pos + 5) !== '.tabs') return false;

        let line = state.getLines(startLine, startLine+1, state.blkIndent, false);
        let lineNum = startLine+1
        let tabs = `<div class="tabs">`
        let tabsHeader = ""
        let tabFooter = ""
        let currentTabContent = ""
        let currentTabTitle = ""
        let currentTabActive = ""
        let currentTabID = Math.random().toString(36).substring(2, 14);

        for(;line!='.tabs.end' && lineNum<endLine; lineNum++){
            line = state.getLines(lineNum, lineNum+1, state.blkIndent, false);
            if (line.startsWith(".tab") ) {
                if (currentTabContent != "") {
                    if (currentTabContent != "") {
                        const content = md.render(currentTabContent);
                        let active = "hidden"
                        if (currentTabActive != "") {
                            active = ""
                        }
                        tabFooter += `<div class="tabcontent `+active+`" id="`+currentTabID+`">`+content+`</div>`;
                        currentTabContent = ""
                        currentTabActive = ""
                        currentTabID = Math.random().toString(36).substring(2, 14);
                    }
                }
                if (line == '.tabs.end') {
                    break
                }
                currentTabTitle = line.slice(5).trim(); // slice(5) to remove ".tab "
                if (line.startsWith(".tab.active")) {
                    currentTabActive = " active"
                    currentTabTitle = line.slice(12).trim();
                }
                tabsHeader = tabsHeader + `<button class="tablinks`+currentTabActive+`" onclick="tabChangeGlobal('`+currentTabID+`')" id='tab-`+currentTabID+`'>`+currentTabTitle+`</button>`
            } else {
                currentTabContent += "\n"+line
            }
        }
        if (tabsHeader != "") {
            tabsHeader = `<div class="tab">` + tabsHeader + `</div>`
            tabs = tabsHeader + `<br>`+ tabs + tabFooter + `</div>`
        }
        if (tabs == `<div class="tabs">`) {
            return false
        }
        if (!silent) {
            const token: Token = state.push('zb_tabs', '', 0);
            token.content = tabs;
        }
        state.line = lineNum + 1;
        return true;        
    });

    md.renderer.rules['zb_tabs'] = function(tokens: Token[], idx: number) {
        return tokens[idx].content;
    };
};
