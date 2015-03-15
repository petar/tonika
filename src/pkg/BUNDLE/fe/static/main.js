var AdminURL = 'http://a.5ttt.org'

function mainReady() {
        $('#social li').hover(
                function() {
                        $(this).animate({ left: "-20px" }, 300);
                },
                function() {
                        $(this).animate({ left: "0px" }, 300);
                }
        );
}
