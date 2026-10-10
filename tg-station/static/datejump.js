/* ============================================================
   按日期快速跳转 —— 交互层
   · 点吸顶日期（.date-header）→ 底部面板：日期输入框 + 日期列表 → 只看某一天
   · 单日模式由主页面管：模块只驱动 window.DateJumpBridge（见 index.html
     「按日期快速跳转：状态与数据层」段），不重复维护 allMedia / allDates。
   · 主页面通过 window.DateJumpUI.syncFocus(date) 通知单日模式开/关。
   ============================================================ */
(function () {
    'use strict';

    var WEEK = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
    var built = false, els = {}, openState = false, suppressUntil = 0, currentFocus = null;

    function bridge() { return window.DateJumpBridge || null; }

    function call(name, arg) {
        var b = bridge();
        if (b && typeof b[name] === 'function') return b[name](arg);
        return null;
    }

    function icon(name) {
        var b = bridge();
        return (b && typeof b.icon === 'function') ? b.icon(name) : '';
    }

    function weekday(date) {
        var t = Date.parse(date + 'T00:00:00');
        return isNaN(t) ? '' : WEEK[new Date(t).getDay()];
    }

    function todayStr() {
        var n = new Date();
        var m = String(n.getMonth() + 1).padStart(2, '0');
        var d = String(n.getDate()).padStart(2, '0');
        return n.getFullYear() + '-' + m + '-' + d;
    }

    // ---------- DOM（懒创建，挂在 body 上） ----------
    function build() {
        if (built) return;

        var mask = document.createElement('div');
        mask.className = 'dj-mask';
        mask.addEventListener('click', close);

        var sheet = document.createElement('div');
        sheet.className = 'dj-sheet';
        sheet.setAttribute('role', 'dialog');
        sheet.setAttribute('aria-modal', 'true');
        sheet.setAttribute('aria-label', '按日期跳转');
        sheet.innerHTML =
            '<div class="dj-head">' +
                '<span class="dj-ic">' + icon('calendar') + '</span>' +
                '<span class="dj-title">按日期跳转</span>' +
                '<button class="dj-close" type="button" aria-label="关闭">' + icon('x') + '</button>' +
            '</div>' +
            '<div class="dj-tools">' +
                '<input class="dj-input" type="date" aria-label="选择日期">' +
                '<button class="dj-btn dj-back" type="button">回到时间线</button>' +
            '</div>' +
            '<div class="dj-list"></div>';

        document.body.appendChild(mask);
        document.body.appendChild(sheet);

        els = {
            mask: mask,
            sheet: sheet,
            list: sheet.querySelector('.dj-list'),
            input: sheet.querySelector('.dj-input'),
            back: sheet.querySelector('.dj-back'),
            close: sheet.querySelector('.dj-close')
        };

        els.close.addEventListener('click', close);
        els.back.addEventListener('click', function () { leaveFocus(); });
        els.input.addEventListener('change', function () {
            if (els.input.value) jump(els.input.value);
        });
        els.list.addEventListener('click', function (e) {
            var row = e.target && e.target.closest ? e.target.closest('.dj-row') : null;
            if (row && row.dataset.date) jump(row.dataset.date);
        });
        document.addEventListener('keydown', function (e) {
            if (e.key === 'Escape' && openState) close();
        });

        built = true;
    }

    // ---------- 日期列表 ----------
    function renderList() {
        if (!built) return;
        var dates = call('dates') || [];
        var focus = call('focusDate');
        var today = todayStr();

        if (!dates.length) {
            els.list.innerHTML = '<div class="dj-empty">日期列表还没准备好，稍后再试</div>';
            return;
        }
        var html = '';
        for (var i = 0; i < dates.length; i++) {
            var d = dates[i];
            var tag = d === today ? '今天' : (d === focus ? '当前' : '');
            html += '<div class="dj-row' + (d === focus ? ' active' : '') + '" data-date="' + d + '">' +
                '<span class="dj-d">' + d + '</span>' +
                '<span class="dj-w">' + weekday(d) + '</span>' +
                (tag ? '<span class="dj-tag">' + tag + '</span>' : '') +
                '</div>';
        }
        els.list.innerHTML = html;
    }

    function scrollActiveIntoView(fallbackDate) {
        var el = els.list.querySelector('.dj-row.active') ||
            (fallbackDate ? els.list.querySelector('.dj-row[data-date="' + fallbackDate + '"]') : null);
        if (el && el.scrollIntoView) el.scrollIntoView({ block: 'center' });
    }

    // ---------- 打开 / 关闭 ----------
    function open(fromDate) {
        build();
        var focus = call('focusDate');
        els.back.style.display = focus ? '' : 'none';
        els.input.value = focus || fromDate || '';
        renderList();
        els.mask.classList.add('show');
        els.sheet.classList.add('show');
        openState = true;
        // 面板动画结束后再滚动，否则 scrollIntoView 在隐藏容器里量不到位置
        setTimeout(function () { scrollActiveIntoView(fromDate); }, 60);
    }

    function close() {
        if (!built) return;
        openState = false;
        els.mask.classList.remove('show');
        els.sheet.classList.remove('show');
    }

    async function jump(date) {
        close();
        // 短抑制窗口：连点/双击时第二次点击不再把面板弹回来
        suppressUntil = Date.now() + 400;
        await call('focus', date);
    }

    async function leaveFocus() {
        close();
        suppressUntil = Date.now() + 400;
        await call('clear');
    }

    // ---------- 单日模式开关：主页面回调用 ----------
    // 不用悬浮条（窄屏会和状态行打架）：直接把「返回时间线」做成吸顶日期
    // 行右侧的一个小胶囊，跟着吸顶行一起走。
    function syncFocus(date) {
        build();
        currentFocus = date || null;
        var olds = document.querySelectorAll('.dj-back-inline');
        for (var i = 0; i < olds.length; i++) olds[i].parentNode.removeChild(olds[i]);
        els.back.style.display = currentFocus ? '' : 'none';
        if (currentFocus) els.input.value = currentFocus;
        markHeaders(document);          // 需要时把胶囊加回去
        if (openState) renderList();
    }

    // ---------- 吸顶日期：加可点提示（+ 单日模式下的返回胶囊） ----------
    function markHeaders(root) {
        var scope = root || document;
        var headers = scope.querySelectorAll('.date-header');
        for (var i = 0; i < headers.length; i++) {
            var h = headers[i];
            var first = h.firstChild;
            var text = (first && first.nodeType === 3 ? first.nodeValue : h.textContent) || '';
            var date = text.trim().slice(0, 10);
            if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) continue;
            h.dataset.djDate = date;

            if (!h.querySelector('.dj-hint')) {
                var hint = document.createElement('span');
                hint.className = 'dj-hint';
                hint.innerHTML = icon('calendar') + '<i></i>';
                h.appendChild(hint);
            }
            // 单日模式：只有被聚焦的那一天挂「返回时间线」
            if (currentFocus && date === currentFocus && !h.querySelector('.dj-back-inline')) {
                var back = document.createElement('button');
                back.type = 'button';
                back.className = 'dj-back-inline';
                back.textContent = '返回时间线';
                back.addEventListener('click', function (ev) {
                    ev.stopPropagation();
                    leaveFocus();
                });
                h.appendChild(back);
            }
        }
    }

    function observeContent() {
        var content = document.getElementById('content');
        if (!content) return;
        var pending = false;
        var obs = new MutationObserver(function () {
            if (pending) return;
            pending = true;
            requestAnimationFrame(function () {
                pending = false;
                markHeaders(content);     // 幂等：已有提示的跳过，不会自激
                if (openState) renderList();
            });
        });
        obs.observe(content, { childList: true, subtree: true });
        markHeaders(content);
    }

    document.addEventListener('click', function (e) {
        if (openState) return;                       // 面板已开，交给面板自己处理
        if (Date.now() < suppressUntil) return;      // 刚选完日期，忽略同一次连点的余波
        var h = e.target && e.target.closest ? e.target.closest('.date-header') : null;
        if (!h) return;
        var b = bridge();
        if (!b || (typeof b.isSelectMode === 'function' && b.isSelectMode())) return;
        if (e.target.closest('.select-hint')) return;
        open(h.dataset.djDate);
    }, true);

    function init() {
        build();
        observeContent();
        syncFocus(call('focusDate'));
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    // 给主页面用（单日模式开关时同步返回条 / 面板）
    window.DateJumpUI = { syncFocus: syncFocus, open: open, close: close };
})();
