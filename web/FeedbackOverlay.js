class FeedbackOverlay {
    constructor() {
        this.feedback = [];
        this.isAdding = false;
        this.init();
    }

    init() {
        const root = document.createElement('div');
        root.id = 'feedback-root';
        document.body.appendChild(root);

        const btn = document.createElement('button');
        btn.textContent = 'Add Feedback';
        btn.style.cssText = 'position: fixed; bottom: 20px; right: 20px; padding: 10px 20px; background: #F39800; color: #fff; border: none; border-radius: 5px; cursor: pointer; font-weight: bold; z-index: 9999; box-shadow: 0 4px 10px rgba(0,0,0,0.2);';
        btn.onclick = () => this.toggleAdding();
        root.appendChild(btn);

        document.addEventListener('click', (e) => this.handleClick(e));
    }

    toggleAdding() {
        this.isAdding = !this.isAdding;
        document.body.style.cursor = this.isAdding ? 'crosshair' : 'default';
        alert(this.isAdding ? 'Click on any element to add feedback.' : 'Feedback mode disabled.');
    }

    handleClick(e) {
        if (!this.isAdding) return;
        if (e.target.tagName === 'BUTTON') return;

        e.preventDefault();
        e.stopPropagation();

        const comment = prompt('Enter your feedback for this element:');
        if (comment) {
            const variant = e.target.closest('[data-variant]')?.getAttribute('data-variant') || 'General';
            this.feedback.push({
                variant,
                element: e.target.tagName,
                text: e.target.textContent.substring(0, 30).trim(),
                comment
            });
        }
        
        this.isAdding = false;
        document.body.style.cursor = 'default';

        if (confirm('Feedback added. Do you want to submit all feedback now?')) {
            this.submit();
        }
    }

    submit() {
        let output = '## Design Lab Feedback\n\n';
        const overall = prompt('Overall Direction (Required):');
        
        this.feedback.forEach((f, i) => {
            output += `### Variant ${f.variant}\n`;
            output += `${i+1}. **${f.element}**\n   "${f.comment}"\n\n`;
        });

        output += `### Overall Direction\n${overall || 'No overall direction provided.'}`;
        
        const textarea = document.createElement('textarea');
        textarea.value = output;
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
        
        alert('Feedback copied to clipboard! Please paste it into the terminal.');
        this.feedback = [];
    }
}

new FeedbackOverlay();
