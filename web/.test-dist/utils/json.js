/**
 * prettyJsonText 将字符串或对象格式化成便于编辑的 JSON 文本。
 * 这里统一做缩进，避免每个页面重复处理回显逻辑。
 */
export function prettyJsonText(value, fallback = '{}') {
    if (value === null || value === undefined || value === '') {
        return fallback;
    }
    if (typeof value === 'string') {
        try {
            return JSON.stringify(JSON.parse(value), null, 2);
        }
        catch {
            return value;
        }
    }
    return JSON.stringify(value, null, 2);
}
/**
 * parseJsonText 在保存前做一次严格校验，并返回标准化对象。
 */
export function parseJsonText(text) {
    return JSON.parse(text);
}
/**
 * normalizeJsonText 将编辑器文本规整为一行 JSON，方便后端直接存储。
 */
export function normalizeJsonText(text) {
    return JSON.stringify(parseJsonText(text));
}
