interface Config {
    counters: {
        letterCount: number;
        langCount: number;
    };
    cursorIsHidden: boolean;
    elements: {
        console: string;
        cursor: string;
    };
    direction: 'ltr' | 'rtl';
    lang: string[];
    speeds: {
        flash: number;
        typeSlow: number;
        typeFast: number;
        delay: number;
    };
    wait: boolean;
    id: number | null;
    typeSpeed: number | null;
    loop: boolean;
}

function run() {
    const config: Config = {
        counters: {
            letterCount: 1,
            langCount: 1,
        },
        cursorIsHidden: true,
        elements: {
            console: "console",
            cursor: "console-text"
        },
        direction: 'ltr',
        lang: ['Hey.', 'You can skip the wiring code.', 'Batteries are included.', 'ADELE'],
        speeds: {
            flash: 400,
            typeSlow: 120,
            typeFast: 70,
            delay: 1000
        },
        wait: false,
        id: null,
        typeSpeed: null,
        loop: false
    }

    const consoleText = document.getElementById(config.elements.cursor);

    function cursorBlink(): void {
        const console = document.getElementById(config.elements.console);
        if (!console) {
            throw new Error('Console not found');
        }
        if (config.cursorIsHidden) {
            console.className = 'console-cursor !inline-block not-italic top-[-4px] left-[10px] opacity-0'
            config.cursorIsHidden = false;
            return;
        }
        console.className = 'console-cursor !inline-block not-italic top-[-4px] left-[10px]'
        config.cursorIsHidden = true;
    }

    function typing(): void {
        if (!config.wait) {
            if (!consoleText) {
                throw new Error('Console text not found');
            }
            consoleText.innerHTML = config.lang[0].substring(0, config.counters.letterCount)
            if (config.direction === 'ltr') {
                config.counters.letterCount++;
            } else if (config.direction === 'rtl') {
                config.counters.letterCount -= 1;
            }
        }
    }

    function atStart(): void {
        if (config.counters.letterCount == 0 && !config.wait) {
            config.wait = true;
            if (!consoleText) {
                throw new Error('Console text not found');
            }
            consoleText.innerHTML = config.lang[0].substring(0, config.counters.letterCount)
            window.setTimeout(function () {
                config.lang.push(config.lang.shift());
                config.direction = 'ltr'
                config.counters.letterCount++;
                config.wait = false;
                clearInterval(config.id as number);
                type();
            }, config.speeds.delay)
        }
    }

    function atEnd(): void {
        if (config.counters.letterCount === config.lang[0].length + 1 && !config.wait) {
            config.wait = true;
            window.setTimeout(function () {
                config.counters.langCount++
                config.direction = 'rtl';
                config.counters.letterCount -= 1;
                config.wait = false;
                clearInterval(config.id as number);
                type();
            }, 1000)
        }
    }

    function type(): void {
        if (!config.loop && config.counters.langCount > config.lang.length) {
            return
        }
        if (config.direction === 'ltr') {
            config.typeSpeed = config.speeds.typeSlow
        } else {
            config.typeSpeed = config.speeds.typeFast
        }
        config.id = window.setInterval(() => {
            atStart();
            atEnd();
            typing();
        }, config.typeSpeed as number);
    }

    window.setInterval(() => {
        cursorBlink()
    }, config.speeds.flash)

    window.setTimeout(() => { type() }, 1000)
}

try {
    run()
} catch (e) {
    console.log(e)
}

