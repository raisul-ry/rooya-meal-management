/* ── Meal checkbox toggle ──────────────────────────── */
document.addEventListener('change', async (e) => {
    const cb = e.target;
    if (!cb.classList.contains('meal-cb')) return;

    const memberId = cb.dataset.member;
    const date     = cb.dataset.date;
    const prev     = !cb.checked;

    cb.disabled = true;

    try {
        const res  = await fetch('/api/meals/toggle', {
            method:  'POST',
            headers: {'Content-Type': 'application/json'},
            body:    JSON.stringify({member_id: memberId, date})
        });
        const data = await res.json();

        if (!res.ok) {
            cb.checked = prev;
            toast(data.error || 'Update failed', true);
        } else {
            const dayEl = document.getElementById(`total-${date}`);
            if (dayEl) dayEl.textContent = data.day_total;
            const mEl = document.querySelector(`[data-m-date="${date}"]`);
            if (mEl) mEl.textContent = data.day_total;
            const chip = cb.closest('.day-chip');
            if (chip) chip.classList.toggle('day-chip-on', cb.checked);
            recalcGrandTotal();
            const mgt = document.getElementById('m-grand-total');
            const gt  = document.getElementById('grand-total');
            if (mgt && gt) mgt.textContent = gt.textContent;
        }
    } catch {
        cb.checked = prev;
        toast('Network error', true);
    } finally {
        cb.disabled = cb.dataset.locked === 'true' ? true : false;
    }
});

function recalcGrandTotal() {
    const sum = [...document.querySelectorAll('.td-day-total')]
        .reduce((acc, el) => acc + (parseInt(el.textContent) || 0), 0);
    const el = document.getElementById('grand-total');
    if (el) el.textContent = sum;
}

/* ── Add Member modal ──────────────────────────────── */
function showAddMemberModal() {
    const modal = document.getElementById('addMemberModal');
    if (!modal) return;
    modal.classList.add('open');
    document.getElementById('memberNameInput').value = '';
    document.getElementById('memberPassword').value = '';
    document.getElementById('memberError').textContent = '';
    setTimeout(() => document.getElementById('memberNameInput').focus(), 60);
}

function hideAddMemberModal() {
    document.getElementById('addMemberModal')?.classList.remove('open');
}

document.getElementById('addMemberModal')?.addEventListener('click', (e) => {
    if (e.target === e.currentTarget) hideAddMemberModal();
});

document.getElementById('memberNameInput')?.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') hideAddMemberModal();
});

async function addMember() {
    const textarea = document.getElementById('memberNameInput');
    const pwInput  = document.getElementById('memberPassword');
    const errEl    = document.getElementById('memberError');

    errEl.textContent = '';
    const names = textarea.value.split('\n').map(n => n.trim()).filter(Boolean);
    if (!names.length) { errEl.textContent = 'Enter at least one name'; return; }

    try {
        const res  = await fetch('/api/members', {
            method:  'POST',
            headers: {'Content-Type': 'application/json'},
            body:    JSON.stringify({names, password: pwInput.value})
        });
        const data = await res.json();

        if (!res.ok) {
            errEl.textContent = data.error || 'Failed to add members';
        } else {
            const msg = data.skipped > 0
                ? `Added ${data.added}, skipped ${data.skipped} duplicate(s)`
                : `Added ${data.added}`;
            hideAddMemberModal();
            toast(msg);
            location.reload();
        }
    } catch {
        errEl.textContent = 'Network error';
    }
}

/* ── Delete Member ─────────────────────────────────── */
let _deleteId = null;

function deleteMember(memberId, name) {
    _deleteId = memberId;
    document.getElementById('deleteConfirmText').textContent =
        `Remove "${name}"? This will also clear their future meal sign-ups.`;
    document.getElementById('deletePassword').value = '';
    document.getElementById('deleteError').textContent = '';
    document.getElementById('deleteMemberModal').classList.add('open');
    setTimeout(() => document.getElementById('deletePassword').focus(), 60);
}

function hideDeleteModal() {
    document.getElementById('deleteMemberModal').classList.remove('open');
    _deleteId = null;
}

document.getElementById('deleteMemberModal')?.addEventListener('click', (e) => {
    if (e.target === e.currentTarget) hideDeleteModal();
});

async function confirmDelete() {
    if (!_deleteId) return;
    const pw    = document.getElementById('deletePassword').value;
    const errEl = document.getElementById('deleteError');
    errEl.textContent = '';

    try {
        const res  = await fetch(`/api/members/${_deleteId}`, {
            method:  'DELETE',
            headers: {'Content-Type': 'application/json'},
            body:    JSON.stringify({password: pw})
        });
        const data = await res.json();
        if (!res.ok) {
            errEl.textContent = data.error || 'Failed to remove member';
        } else {
            hideDeleteModal();
            location.reload();
        }
    } catch {
        errEl.textContent = 'Network error';
    }
}

/* ── Settings form ─────────────────────────────────── */
function saveSettings(e) {
    e.preventDefault();
    const form = e.target;
    fetch('/settings', {method: 'POST', body: new FormData(form)})
        .then(r => r.json())
        .then(data => {
            if (data.success) {
                const msg = document.getElementById('saveMsg');
                msg.classList.add('visible');
                setTimeout(() => msg.classList.remove('visible'), 2200);
                form.password.value = '';
            } else {
                toast(data.error || 'Save failed', true);
            }
        })
        .catch(() => toast('Save failed', true));
}

/* ── Menu PDF upload ───────────────────────────────── */
function uploadMenu(input) {
    const file = input.files[0];
    if (!file) return;

    const fd = new FormData();
    fd.append('pdf', file);

    fetch('/api/menu/upload', {method: 'POST', body: fd})
        .then(r => r.json())
        .then(data => {
            if (data.success) location.reload();
            else toast(data.error || 'Upload failed', true);
        })
        .catch(() => toast('Network error', true));
}

/* ── Auto-refresh at midnight ──────────────────────── */
(function() {
    const now = new Date();
    const midnight = new Date(now);
    midnight.setHours(24, 0, 0, 0);
    setTimeout(() => location.reload(), midnight - now);
})();

/* ── Side nav (mobile) ─────────────────────────────── */
function openSideNav() {
    document.getElementById('sidenav').classList.add('open');
    document.getElementById('sidenavOverlay').classList.add('open');
}

function closeSideNav() {
    document.getElementById('sidenav').classList.remove('open');
    document.getElementById('sidenavOverlay').classList.remove('open');
}

/* ── Toast helper ──────────────────────────────────── */
function toast(msg, isErr = false) {
    const el = document.getElementById('toast');
    if (!el) return;
    el.textContent = msg;
    el.className   = `toast show${isErr ? ' err' : ''}`;
    clearTimeout(el._t);
    el._t = setTimeout(() => { el.className = 'toast'; }, 3200);
}
