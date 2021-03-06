(function ( $ ) {
 
    $.fn.rebind = function() {
        $(this).click(function() {
            $(this).loadTransactions();
            return false;
        });
    };

    $.fn.loadTransactions = function() {
        var loaderHTML = `
        <div class="loader"><div>
            <svg width="30" height="30" viewBox="0 0 135 135" xmlns="http://www.w3.org/2000/svg" fill="#FF6700">
                <path d="M67.447 58c5.523 0 10-4.477 10-10s-4.477-10-10-10-10 4.477-10 10 4.477 10 10 10zm9.448 9.447c0 5.523 4.477 10 10 10 5.522 0 10-4.477 10-10s-4.478-10-10-10c-5.523 0-10 4.477-10 10zm-9.448 9.448c-5.523 0-10 4.477-10 10 0 5.522 4.477 10 10 10s10-4.478 10-10c0-5.523-4.477-10-10-10zM58 67.447c0-5.523-4.477-10-10-10s-10 4.477-10 10 4.477 10 10 10 10-4.477 10-10z">
                    <animateTransform
                        attributeName="transform"
                        type="rotate"
                        from="0 67 67"
                        to="-360 67 67"
                        dur="2.5s"
                        repeatCount="indefinite"/>
                </path>
                <path d="M28.19 40.31c6.627 0 12-5.374 12-12 0-6.628-5.373-12-12-12-6.628 0-12 5.372-12 12 0 6.626 5.372 12 12 12zm30.72-19.825c4.686 4.687 12.284 4.687 16.97 0 4.686-4.686 4.686-12.284 0-16.97-4.686-4.687-12.284-4.687-16.97 0-4.687 4.686-4.687 12.284 0 16.97zm35.74 7.705c0 6.627 5.37 12 12 12 6.626 0 12-5.373 12-12 0-6.628-5.374-12-12-12-6.63 0-12 5.372-12 12zm19.822 30.72c-4.686 4.686-4.686 12.284 0 16.97 4.687 4.686 12.285 4.686 16.97 0 4.687-4.686 4.687-12.284 0-16.97-4.685-4.687-12.283-4.687-16.97 0zm-7.704 35.74c-6.627 0-12 5.37-12 12 0 6.626 5.373 12 12 12s12-5.374 12-12c0-6.63-5.373-12-12-12zm-30.72 19.822c-4.686-4.686-12.284-4.686-16.97 0-4.686 4.687-4.686 12.285 0 16.97 4.686 4.687 12.284 4.687 16.97 0 4.687-4.685 4.687-12.283 0-16.97zm-35.74-7.704c0-6.627-5.372-12-12-12-6.626 0-12 5.373-12 12s5.374 12 12 12c6.628 0 12-5.373 12-12zm-19.823-30.72c4.687-4.686 4.687-12.284 0-16.97-4.686-4.686-12.284-4.686-16.97 0-4.687 4.686-4.687 12.284 0 16.97 4.686 4.687 12.284 4.687 16.97 0z">
                    <animateTransform
                        attributeName="transform"
                        type="rotate"
                        from="0 67 67"
                        to="360 67 67"
                        dur="8s"
                        repeatCount="indefinite"/>
                </path>
            </svg>
        </div></div>`;
        var link = this;
        var url = link.attr('data-url');
        var id = link.attr('id');
        var rowId = 'invoice-' + id;
        var rowHTML = `
        <tr id="${rowId}-transactions" height="100px">
            <td>
                <a href="#" class="transactions-loader" data-url="${url}">
                    <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 32 32" version="1.1">
                        <g id="surface1">
                            <path style=" " d="M 16 4 C 10.886719 4 6.617188 7.160156 4.875 11.625 L 6.71875 12.375 C 8.175781 8.640625 11.710938 6 16 6 C 19.242188 6 22.132813 7.589844 23.9375 10 L 20 10 L 20 12 L 27 12 L 27 5 L 25 5 L 25 8.09375 C 22.808594 5.582031 19.570313 4 16 4 Z M 25.28125 19.625 C 23.824219 23.359375 20.289063 26 16 26 C 12.722656 26 9.84375 24.386719 8.03125 22 L 12 22 L 12 20 L 5 20 L 5 27 L 7 27 L 7 23.90625 C 9.1875 26.386719 12.394531 28 16 28 C 21.113281 28 25.382813 24.839844 27.125 20.375 Z "/>
                        </g>
                    </svg>
                </a>
            </td>
            <td colspan="7" style="position:relative;">${loaderHTML}</td>
        </tr>`;
    
        $.ajax({
            url: url,
            beforeSend: function() {
                if ($(`tr#${rowId}-transactions`).length == 0) {
                    $(rowHTML).insertAfter(`#${rowId}`);
                    //$('tr#' + rowId + '-transactions td:eq(1)').prepend(loader);
                } else {
                    $(`tr#${rowId}-transactions td:eq(1)`).children('.loader').show();
                }
            },
            success: function(data) {
                //var pos = $(`tr#${rowId}-transactions td:eq(0) svg`).position();
                var td = $(`tr#${rowId}-transactions td:eq(1)`);
                var loader = td.children('.loader').clone();
                loader.wrapAll('<div>');
                td.html(loader.parent().html() + data);
            },
            error: function (jqXHR, textStatus, errorThrown) {
                //console.log("TEXT:", textStatus);        
            },
            complete: function () {
                $(`tr#${rowId}-transactions td:eq(1)`).children('.loader').hide();
                $(`tr#${rowId}-transactions td:eq(0) a`).rebind();
                feather.replace();
            }
        });
    
        return this;
    };
 
}( jQuery ));

$(document).ready(function () {
    $('.transactions-loader').click(function() {
        var id = $(this).attr('id');
        var transactionsRowId = `#invoice-${id}-transactions`;
        if ($(transactionsRowId).length == 0) {
            $(this).loadTransactions();
            $(this).removeClass('more').addClass('less');
            $('a#' + id).html('<span class="text-danger" data-feather="minus-square" aria-hidden="true"></span>');
        } else if ($(this).hasClass('less')) {
            $(transactionsRowId).hide()
            $(this).removeClass('less').addClass('more');
            $('a#' + id).html('<span class="text-danger" data-feather="plus-square" aria-hidden="true"></span>');
        } else if ($(this).hasClass('more')) {
            $(transactionsRowId).show()
            $(this).removeClass('more').addClass('less');
            $('a#' + id).html('<span class="text-danger" data-feather="minus-square" aria-hidden="true"></span>');
        }
        feather.replace();
        return false;
    });
});