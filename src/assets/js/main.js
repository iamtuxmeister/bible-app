var allowHighlighting = false;
var buttonClicked = "";
var highlightColor = "";
const highlightColors = [
    "bg-red-600",
    "bg-orange-600",
    "bg-amber-600",
    "bg-yellow-600",
    "bg-lime-600",
    "bg-green-600",
    "bg-emerald-600",
    "bg-teal-600",
    "bg-cyan-600",
    "bg-sky-600",
    "bg-rose-600",
    "bg-indigo-600",
    "bg-violet-600",
    "bg-fuchsia-600",
    "bg-pink-600",
    "bg-blue-600"
];

// Highlighted Verses, tuple of verse number and color
var highlightedVerses = {
    "v43003017": "bg-lime-600",
    "v43003019": "bg-blue-600",
    "v43003021": "bg-pink-600"
};

var themeToggleDarkIcon = document.getElementById('theme-toggle-dark-icon');
var themeToggleLightIcon = document.getElementById('theme-toggle-light-icon');

// Change the icons inside the button based on previous settings
if (localStorage.getItem('color-theme') === 'dark' || (!('color-theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    themeToggleLightIcon.classList.remove('hidden');
    document.getElementById("theme-toggle").checked = true;
} else {
    themeToggleDarkIcon.classList.remove('hidden');
    document.getElementById("theme-toggle").checked = false;
}

var themeToggleBtn = document.getElementById('theme-toggle');

themeToggleBtn.addEventListener('click', function() {

    // toggle icons inside button
    themeToggleDarkIcon.classList.toggle('hidden');
    themeToggleLightIcon.classList.toggle('hidden');

    // if set via local storage previously
    if (localStorage.getItem('color-theme')) {
        if (localStorage.getItem('color-theme') === 'light') {
            document.documentElement.classList.add('dark');
            localStorage.setItem('color-theme', 'dark');
        } else {
            document.documentElement.classList.remove('dark');
            localStorage.setItem('color-theme', 'light');
        }

        // if NOT set via local storage previously
    } else {
        if (document.documentElement.classList.contains('dark')) {
            document.documentElement.classList.remove('dark');
            localStorage.setItem('color-theme', 'light');
        } else {
            document.documentElement.classList.add('dark');
            localStorage.setItem('color-theme', 'dark');
        }
    }

});


const searchHistory = new Set();

function addHistoryItem(verse) {
    document.getElementById("history").innerHTML += "<div class=\"history-item\">" +
        "<button class=\"cursor-pointer underline\" onclick=\"useHistory('" + verse + "')\" >" + verse + "</button>"
        + "</div>";
}

function createHistory() {
    document.getElementById("history").innerHTML = "";
    searchHistory.forEach(addHistoryItem);
}

function useHistory(verse) {
    document.getElementById("search").value = verse;
    verseLookup();
}

function clearSearchHistory() {
    document.getElementById('history').innerHTML = '';
    searchHistory.clear();
}

async function verseLookup() {
    var verse = document.getElementById("search").value;
    var headings = document.getElementById("headings").checked;
    var extras = document.getElementById("extras").checked;
    var numbers = document.getElementById("numbers").checked;

    if (verse.match(/(\d+)/) || verse === "") {
        var url = "/api?verse=" + verse + "&headings=" + headings + "&extras=" + extras + "&numbers=" + numbers;
        fetch(url)
            .then(response => response.json())
            .then(data => {
                const retVar = data.passages.join("");
                //console.log(retVar);
                document.getElementById("verse").innerHTML = retVar;
                searchHistory.add(data.query);
                createHistory();
                wrapText();
            });
    } else if (verse.match(/romans road/i) || verse.match(/roman's road/i)) {
        document.getElementById("verse").innerHTML = "<h1>The Romans Road to Salvation</h1>";
        let verses = [
            "Romans 3:23",
            "Romans 3:12",
            "Romans 5:10",
            "Romans 6:23",
            "Romans 5:8",
            "Romans 10:9-10",
            "Romans 10:13",
            "Romans 10:17"
        ];
        let collector = "";
        getDataInOrder(verses).then(data => {
            let count = data.length;
            data.forEach(obj => {
                collector += obj.passages.join("");
                if (--count === 0) {
                    document.getElementById("verse").innerHTML += collector;
                    searchHistory.add("Romans Road");
                    createHistory();
                    wrapText();
                }
            });
        });
    } else {
        var url = "/search?search=" + verse;
        fetch(url)
            .then(response => response.json())
            .then(data => {
                let html = '<ul>';
                data.results.forEach(obj => {
                    html += "<li><button class=\"cursor-pointer underline\" onclick=\"useHistory('" + obj.reference + "')\" >" + obj.reference + "</button>"
                    html += ` -- ${obj.content}</li>`;
                });
                html += '</ul>';
                document.getElementById("verse").innerHTML = html;
                searchHistory.add(verse);
                createHistory();
            });
    }
    window.scrollTo(0, 0);
}


async function getDataInOrder(verses) {
    const promises = verses.map(async (verse) => {
        let url = "/api?verse=" + verse + "&headings=false&extras=false&numbers=false";
        return fetch(url).then(response => response.json());
    });
    return await Promise.all(promises);
}


var inputField = document.getElementById('search');
inputField.addEventListener('keydown', function(event) {
    if (event.key === 'Enter') {
        verseLookup();
    }
});

["headings", "extras", "numbers"].forEach(function(id) {
    document.getElementById(id).addEventListener('change', verseLookup);
});


function wrapText() {
    document.querySelectorAll("span.verse").forEach((span) => {
        const verseID = span.getAttribute("data-verse");
        span.addEventListener("click", (e) => {
            // Read the verse ID from the data-verse attribute
            // Assume highlightColor is set elsewhere in your script
            highlightVerseByID(verseID, highlightColor);
        });
        // Check if the verse is in the highlightedVerses array
        if (highlightedVerses[verseID]) {
            //console.log("Highlighting verse:", verseID);
            if (allowHighlighting) {
                highlightWrapper(span, highlightedVerses[verseID]);
            } else {
                allowHighlighting = true;
                highlightWrapper(span, highlightedVerses[verseID]);
                allowHighlighting = false;
            }
        }
    });
}


/**
 * Finds all <span class="verse" data-verse="..."> elements matching the given verseID
 * and applies highlightWrapper to each.
 *
 * @param {string} verseID        The verse ID to look for (e.g. "v45003010").
 * @param {string} highlightColor The CSS class name that highlightWrapper should apply.
 */
function highlightVerseByID(verseID, highlightColor) {
    // Construct a selector for all <span class="verse" data-verse="verseID">
    const selector = `span.verse[data-verse="${verseID}"]`;
    const wrappers = document.querySelectorAll(selector);

    wrappers.forEach((wrapper) => {
        highlightWrapper(wrapper, highlightColor);
    });

}


function highlightWrapper(wrapper, highlightColor) {
    const children = wrapper.children;
    const verseID = wrapper.getAttribute("data-verse");
    if (allowHighlighting) {
        if (hasClass(wrapper, highlightColor)) {
            removeClass(wrapper, highlightColor);
            for (child of children) {
                if (hasClass(child, "woc-highlighted")) {
                    removeClass(child, "woc-highlighted");
                    addClass(child, "woc");
                }
            }
            delete highlightedVerses[verseID];
        } else {
            highlightColors.forEach((color) => {
                removeClass(wrapper, color);
            });
            addClass(wrapper, highlightColor);
            for (child of children) {
                if (hasClass(child, "woc")) {
                    removeClass(child, "woc");
                    addClass(child, "woc-highlighted");
                }
            }
            highlightedVerses[verseID] = highlightColor;
        }
    }
}

function HighlightButtonClicked(me) {
    //console.log(me.id);
    var buttons = document.querySelectorAll(".highlight-button");
    buttons.forEach(button => {
        removeClass(button, "dark:border-stone-300");
        removeClass(button, "border-violet-800");
        addClass(button, "border-transparent");
    });
    if (allowHighlighting && me.id === buttonClicked) {
        allowHighlighting = false;
    } else {
        removeClass(me, "border-transparent");
        addClass(me, "dark:border-stone-300");
        addClass(me, "border-violet-800");

        buttonClicked = me.id;
        allowHighlighting = true;
        highlightColor = "bg-" + me.id.split("-")[1] + "-600";
        //console.log(highlightColor);

    }
}


function hasClass(ele, cls) {
    return !!ele.className.match(new RegExp('(\\s|^)' + cls + '(\\s|$)'));
}

function addClass(ele, cls) {
    if (!hasClass(ele, cls)) ele.className += " " + cls;
}

function removeClass(ele, cls) {
    if (hasClass(ele, cls)) {
        var reg = new RegExp('(\\s|^)' + cls + '(\\s|$)');
        ele.className = ele.className.replace(reg, ' ');
    }
}

//Add event from js the keep the marup clean
function init() {
    document.getElementById("open-menu").addEventListener("click", toggleMenu);
    document.getElementById("body-overlay").addEventListener("click", toggleMenu);
}

//The actual fuction
function toggleMenu() {
    var ele = document.getElementsByTagName('body')[0];
    if (!hasClass(ele, "menu-open")) {
        addClass(ele, "menu-open");
    } else {
        removeClass(ele, "menu-open");
    }
}

//Prevent the function to run before the document is loaded
document.addEventListener('readystatechange', function() {
    if (document.readyState === "complete") {
        init();
    }
});
