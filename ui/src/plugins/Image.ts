// zbImage.ts
import MarkdownIt from 'markdown-it';
import StateInline from 'markdown-it/lib/rules_inline/state_inline';
import Token from 'markdown-it/lib/token';

export default function zbImage(md: MarkdownIt) {
    md.inline.ruler.after('text', 'zb_image', function(state: StateInline, silent: boolean) {
        const src = state.src;
        const pos = state.pos;

        if (src[pos] !== ':') return false;
        const start = pos + 1;
        const end = src.indexOf(')', start);

        if (end === -1) return false;
        const match = src.slice(start, end);

        const extra = match.split('(');
        if (extra.length < 2) {
            return false
        }

        // Add a regular expression test for the ".image(url opt)" format
        const re = /\image\(([^ ]*)(?: (.*))?/;
        const regMatch = match.match(re);
        if (!regMatch) return false;
        if (!silent) {
            const token: Token = state.push('zb_image', '', 0);
            let opt = `style=" object-fit: contain;`
            //see if we have extra options
            if (regMatch.length > 1 && regMatch[2]) {
                const matchSize = regMatch[2].split(' ');
                if (matchSize.length > 0) {
                    let normalWH = matchSize[0].split(':')
                    if (normalWH.length == 1) {
                        opt += ` width: `+normalWH[0]+`; height:`+normalWH+`;`

                    }
                    if (normalWH.length > 1) {
                        if (normalWH[0] == ""){
                            normalWH[0] = "auto"
                        }
                        opt += ` width: `+normalWH[0]+`;`
                        if (normalWH[1] == ""){
                            normalWH[1] = "auto"
                        }
                        opt += ` height: `+normalWH[1]+`;`
                    }
                }
                if (matchSize.length > 1) {
                    let maxWH = matchSize[1].split(':')
                    if (maxWH.length == 1) {
                        opt += ` max-width: `+maxWH[0]+`; max-height:`+maxWH+`;`

                    }
                    if (maxWH.length > 1) {
                        if (maxWH[0] == ""){
                            maxWH[0] = "auto"
                        }
                        opt += ` max-width: `+maxWH[0]+`;`
                        if (maxWH[1] == ""){
                            maxWH[1] = "auto"
                        }
                        opt += ` max-height: `+maxWH[1]+`;`
                    }
                }
            }
            opt+=`""`

            //<img src="assets/img.png" alt="back" height="400">
            token.content = `<img src="`+regMatch[1]+`"`+opt+`>`;
            token.markup = `:${match}:`;
        }

        state.pos = end + 1;
        return true;
    });

    md.renderer.rules['zb_image'] = function(tokens: Token[], idx: number) {
        return tokens[idx].content;
    };
};
