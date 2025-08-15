function Prompt(){
    let toast = function(c) {

        // переменные находящиеся внутри объекта "c" который будут переопределенны в случае если у нас они в этом объекте не указанны
        const {
            msg = "",
            icon = "success",
            position = "top-end",
        } = c

        const Toast = Swal.mixin({
            toast: true,
            title: msg,
            position: position,
            icon: icon,
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            didOpen: (toast) => {
                toast.onmouseenter = Swal.stopTimer;
                toast.onmouseleave = Swal.resumeTimer;
            }
        });

        Toast.fire({})
    }

    let success = function(c) {
        const {
            msg = "",
            title = "",
            footer = "",
        } = c;

        Swal.fire({
            icon: "success",
            title: title,
            text: msg,
            footer: footer
        });
    }

    let error = function(c) {
        const {
            msg = "",
            title = "",
            footer = "",
        } = c;

        Swal.fire({
            icon: "error",
            title: title,
            text: msg,
            footer: footer
        });
    }

    // async Объявляет асинхронную функцию custom, которая может использовать await.
    async function custom(c) {
        // Это деструктуризация объекта c, которая извлекает свойства msg и title. Однако объект c может содержать и другие свойства, например, callback.
        const {
            icon = "",
            msg = "",
            title = "",
            showConfirmButton = true
        } = c;

        // Извлекает значение свойства value из объекта, возвращенного Swal.fire, и присваивает его переменной formValues.
        // await Ожидает, пока пользователь закроет модальное окно SweetAlert2. Возвращает объект, содержащий результат взаимодействия пользователя.
        // Функция ждёт, пока пользователь закроет модальное окно (await Swal.fire()), не блокируя основной поток.
        const { value: result } = await Swal.fire({
            icon: icon,
            title: title,
            html: msg,
            backdrop: false,
            heightAuto: false,
            width: '600px',
            focusConfirm: false,
            showCancelButton: true,
            showConfirmButton: showConfirmButton,
            // Если в объекте c был передан кастомный didOpen, он выполнится. Если нет — ничего не произойдет.
            didOpen: () => {
                if (c.didOpen !== undefined){
                    c.didOpen()
                }
            },
            /*preConfirm: () => {
                return [
                    document.getElementById("start").value,
                    document.getElementById("end").value
                ];
            }*/
        });
        // Если пользователь подтвердил выбор, отображаем результат
        // Проверяет, существует ли result. Это объект, возвращаемый SweetAlert2 после закрытия модального окна.
        if (result) {
            // Проверяет, не была ли нажата кнопка отмены (например, пользователь закрыл модальное окно, нажав "Cancel" или на крестик).
            if (result.dismiss !== Swal.DismissReason.cancel) {
                // Проверяет, не пустое ли значение result.value. Это значение возвращается из модального окна (например, данные из формы).
                if (result[0] !== "" && result[1] !== "") {
                    // Проверяет, существует ли callback в объекте c. Объект c — это аргумент, переданный в функцию custom.
                    // Если callback передан, он вызывается с аргументом result.
                    if ( c.callback !== undefined) {
                        c.callback(result);
                    }
                    // Если result.value пустое, вызывается callback с аргументом false. Это может быть полезно, чтобы указать, что данные не были введены.
                } else {
                    c.callback(false);
                }
                // Если пользователь нажал "Cancel" или закрыл модальное окно, вызывается callback с аргументом false.
            } else {
                c.callback(false);
            }
        }

    }

    return {
        toast: toast,
        success: success,
        error: error,
        custom: custom,
    }
}