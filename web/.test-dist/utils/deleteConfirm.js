/**
 * buildDeleteConfirmContent 统一生成删除确认文案。
 * 这样各管理页可以复用同一份提示风格，避免资源名和影响说明写散后不一致。
 */
export function buildDeleteConfirmContent(resourceName, identifier, impact) {
    const lines = [`确认删除${resourceName}“${identifier}”吗？`];
    if (impact) {
        lines.push(impact);
    }
    lines.push('该操作不可撤销。');
    return lines.join(' ');
}
