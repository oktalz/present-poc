// zbFontAwesome.ts
import MarkdownIt from 'markdown-it';
import StateInline from 'markdown-it/lib/rules_inline/state_inline';
import Token from 'markdown-it/lib/token';

export default function zbFontAwesome(md: MarkdownIt) {
    md.inline.ruler.after('text', 'font_awesome', function(state: StateInline, silent: boolean) {
        const src = state.src;
        const pos = state.pos;

        if (src[pos] !== ':') return false;
        const start = pos + 1;
        const end = src.indexOf(':', start);

        if (end === -1) return false;
        const match = src.slice(start, end);

        const extra = match.split('#');
        let extraClass = ""
        if (extra.length > 1) {
            extraClass = extra[1]
        }

        const split = extra[0].split('-');
        if (split.length < 2) return false;

        // Add a regular expression test for the "fa${split[0]}-${split[1]}#extra" format
        const re = /^fa[0-9a-zA-Z]*(-[0-9a-zA-Z]*)+(#[0-9a-zA-Z- ]*)?$/;
        if (!re.test(match)) return false;

        if (!silent) {
            const token: Token = state.push('font_awesome_inline', '', 0);
            token.content = `<i class="${split[0]} fa-${split.slice(1).join('-')} ${extraClass}"></i>`;
            token.markup = `:${match}:`;
        }

        state.pos = end + 1;
        return true;
    });

    md.renderer.rules['font_awesome_inline'] = function(tokens: Token[], idx: number) {
        return tokens[idx].content;
    };
};
