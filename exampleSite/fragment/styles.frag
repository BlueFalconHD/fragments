this:addBuilders {
    styles = function()
        -- Users can override this builder to inject extra CSS, or set shared meta `extraStyles`
        local styles = this:getSharedMeta("extraStyles")
        if styles == nil or tostring(styles) == "nil" or tostring(styles) == "" then
            return ""
        end
        return "<style>" .. tostring(styles) .. "</style>"
    end
}

~~~
*{styles}
